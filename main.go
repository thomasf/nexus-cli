package main

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"

	"github.com/hanjos/nexus"
	"github.com/hanjos/nexus/credentials"
	"github.com/hanjos/nexus/search"
	version "github.com/hashicorp/go-version"
	flags "github.com/jessevdk/go-flags"
	"github.com/thomasf/lg"
)

type Options struct {
	User         string `long:"user" description:"username" ini-name:"username"`
	Password     string `long:"password" description:"password" ini-name:"password"`
	Host         string `long:"host" description:"nexus url" ini-name:"host"`
}

type FilterOptions struct {
	Snapshot bool `long:"snapshot" description:"Only show snapshots"`
	Release  bool `long:"release" description:"Only show releases" `
	Latest   bool `long:"latest" description:"Only show latest (semver) versions, implies --release"`
	POM      bool `long:"pom" description:"Show pom results"`
}

func (f *FilterOptions) Filter(artifacts []Artifact) []Artifact {
	if !f.POM {
		artifacts = filterPOM(artifacts)
	}
	if f.Latest {
		artifacts = getLatest(artifacts)
		sort.Sort(BySemver(artifacts))

	}
	if f.Release {
		artifacts = getReleases(artifacts)
		sort.Sort(BySemver(artifacts))
	}
	if f.Snapshot {
		artifacts = getSnapshots(artifacts)

	}
	return artifacts
}

// // LoggingOptions .
// type LoggingOptions struct {
// 	LogToMemory bool
// 	flag.BoolVar(&logging.toMemory, "logtomemory", false, "log to memory")
// 	flag.BoolVar(&logging.toFile, "logtofile", true, "log to files")
// 	flag.BoolVar(&logging.toStderr, "logtostderr", false, "log to standard error")
// 	flag.BoolVar(&logging.color, "logcolor", false, "use colors in standard error display")

// 	flag.Var(&logging.verbosity, "v", "log level for V logs")
// 	flag.Var(&logging.stderrThreshold, "stderrthreshold", "logs at or above this threshold go to stderr")
// 	flag.Var(&logging.vmodule, "vmodule", "comma-separated list of pattern=N settings for file-filtered logging")
// 	flag.Var(&logging.traceLocation, "log_backtrace_at", "when logging hits line file:N, emit a stack trace")

// }

var options Options

var parser *flags.Parser

func main() {
	flag.Set("logtostderr", "true")
	flag.Set("logcolor", "true")
	flag.Set("v", "100")
	flag.CommandLine.Parse([]string{})

	lg.SetSrcHighlight("thomasf/nexus-cli")
	parser = flags.NewParser(&options, flags.Default)
	err := flags.IniParse("settings.ini", &options)
	if err != nil {
		log.Fatal(err)
	}

	parser.AddCommand("search", "search repo", "", &searchCommand)
	parser.AddCommand("get", "get artifact(s)", "", &getCommand)
	if _, err := parser.Parse(); err != nil {
		os.Exit(1)
	}
}

func getReleases(artifacts []Artifact) []Artifact {
	var result []Artifact

	for _, v := range artifacts {
		if !strings.HasSuffix(v.Version, "SNAPSHOT") {
			if v.v == nil {
				lg.Warningln("could not parse version from", v)
				continue
			}

			result = append(result, v)
		}
	}
	return result
}

func getSnapshots(artifacts []Artifact) []Artifact {
	var result []Artifact
	for _, v := range artifacts {
		if strings.HasSuffix(v.Version, "SNAPSHOT") {
			result = append(result, v)
		}

	}
	return result
}

func getLatest(artifacts []Artifact) []Artifact {
	var result []Artifact

	byArtifact := make(map[string][]Artifact, 0)
	for _, a := range getReleases(artifacts) {
		if !strings.HasSuffix(a.Version, "SNAPSHOT") {
			parts := []string{a.GroupID, a.ArtifactID, a.Extension, a.Classifier, a.RepositoryID}
			key := strings.Join(parts, "\u0000")
			byArtifact[key] = append(byArtifact[key], a)
		}
	}
	for _, v := range byArtifact {
		sort.Sort(BySemver(v))
		result = append(result, v[len(v)-1])
	}
	return result
}

func filterPOM(artifacts []Artifact) []Artifact {
	var result []Artifact
	for _, v := range artifacts {
		if v.Extension != "pom" {
			result = append(result, v)
		}
	}
	return result
}

type SearchCommand struct {
	FilterOptions FilterOptions
}

func (s *SearchCommand) Execute(args []string) error {
	if len(args) < 0 {
		searchrepo("")
		return nil
	}
	artifacts, err := searchrepo(args[0])
	if err != nil {
		return err
	}

	artifacts = s.FilterOptions.Filter(artifacts)

	for _, v := range artifacts {
		fmt.Println(v)
	}

	return nil
}

var searchCommand SearchCommand

type GetCommand struct {
	Output        string `long:"out" short:"o" description:"output path"`
	FilterOptions FilterOptions
}

func (s *GetCommand) Execute(args []string) error {

	gav := args[0]

	creds := credentials.BasicAuth(options.User, options.Password)
	n := nexus.New(options.Host, creds)

	arts, err := n.Artifacts(ParseGAV(gav))
	if err != nil {
		lg.Fatal(err)
	}
	var artifacts []Artifact
	for _, a := range arts {
		art, err := newArtifact(a)
		if err != nil {
			return err
		}
		artifacts = append(artifacts, art)
	}
	artifacts = s.FilterOptions.Filter(artifacts)

	lg.Infoln(artifacts)
	if len(artifacts) < 1 {
		return errors.New("no matching artifact")
	}

	if s.Output != "" {
		if len(artifacts) != 1 {
			return errors.New("cannot use --out with multiple matches")
		}
		v := artifacts[0]
		info, err := n.InfoOf(v.Artifact)
		if err != nil {
			lg.Fatal(err)
		}
		lg.Infoln(info.URL)
		err = download(s.Output, info.URL, creds)
		if err != nil {
			lg.Fatal(err)
		}
		return nil
	}

	for _, v := range artifacts {
		info, err := n.InfoOf(v.Artifact)
		if err != nil {
			lg.Fatal(err)
		}
		lg.Infoln(info.URL)

		dirname := filepath.Join(v.GroupID, v.ArtifactID, v.Version)
		filename := filepath.Base(info.URL)
		err = os.MkdirAll(dirname, 0775)
		if err != nil {
			lg.Fatal(err)
		}

		err = download(filepath.Join(dirname, filename), info.URL, creds)
		if err != nil {
			lg.Fatal(err)
		}

	}

	return nil

}

var getCommand GetCommand


func searchrepo(q string) ([]Artifact, error) {
	var results []Artifact
	// n := nexus.New(options.Host, credentials.None)
	n := nexus.New(options.Host, credentials.BasicAuth(options.User, options.Password))

	var RepositoryID string
	if pos := strings.LastIndex(q, "@"); pos != -1 {
		RepositoryID = q[pos+1:]
		q = q[:pos]
	}

	var crit search.Criteria

	if q == "" {
		crit = search.All
	} else if RepositoryID != "" {
		crit = search.InRepository{
			RepositoryID: RepositoryID,
			Criteria:     search.ByKeyword(q),
		}
	} else {
		crit = search.ByKeyword(q)
	}

	artifacts, err := n.Artifacts(
		crit,
	)
	if err != nil {
		fmt.Printf("%v: %v", reflect.TypeOf(err), err)
		return results, nil
	}
	for _, a := range artifacts {
		art, err := newArtifact(a)
		if err != nil {
			return (make([]Artifact, 0)), err
		}
		results = append(results, art)
	}
	return results, nil
}

// String implements the fmt.Stringer interface, as per Maven docs
// (http://maven.apache.org/pom.html#Maven_Coordinates).

func ParseGAV(gav string) search.Criteria {
	var RepositoryID string
	if pos := strings.LastIndex(gav, "@"); pos != -1 {
		RepositoryID = gav[pos+1:]
		gav = gav[:pos]
	}
	parts := strings.Split(gav, ":")
	var coords search.Criteria
	// this is stupid, fix it later
	switch len(parts) {
	case 3:
		coords = search.ByCoordinates{
			GroupID:    parts[0],
			ArtifactID: parts[1],
			Version:    parts[2],
		}
	case 4:
		coords = search.ByCoordinates{
			GroupID:    parts[0],
			ArtifactID: parts[1],
			Packaging:  parts[2],
			Version:    parts[3],
		}
	case 5:
		coords = search.ByCoordinates{
			GroupID:    parts[0],
			ArtifactID: parts[1],
			Packaging:  parts[2],
			Classifier: parts[3],
			Version:    parts[4],
		}
	default:
		lg.Fatal("invalid string")
	}

	if RepositoryID != "" {
		c := search.InRepository{
			RepositoryID: RepositoryID,
			Criteria:     coords,
		}
		lg.Infoln(c)
		return c
	}
	return coords
}

type BySemver []Artifact

func (v BySemver) Len() int {
	return len(v)
}

func (v BySemver) Less(i, j int) bool {
	return v[i].LessThan(v[j])
}

func (v BySemver) Swap(i, j int) {
	v[i], v[j] = v[j], v[i]
}

// GreaterThan compares the semver version number
func (a *Artifact) GreaterThan(b Artifact) bool {
	return a.v.GreaterThan(b.v)
}

// LessThan compares the semver version number
func (a *Artifact) LessThan(b Artifact) bool {
	return a.v.LessThan(b.v)
}

// Equal compares the semver version number, nothing else.
func (a *Artifact) Equal(b Artifact) bool {
	return a.v.Equal(b.v)
}

// Artifact combines nexus artifact with semver features
type Artifact struct {
	*nexus.Artifact
	v *version.Version
}

func newArtifact(a *nexus.Artifact) (Artifact, error) {
	v, _ := version.NewVersion(a.Version)
	// if err != nil {
	// return nil, err

	// }
	return Artifact{
		Artifact: a,
		v:        v,
	}, nil
}

func download(dst, url string, creds credentials.Credentials) error {
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()
	req, err := http.NewRequest("GET", url, bytes.NewBufferString(url))
	if err != nil {
		lg.Fatal(err)
	}
	creds.Sign(req)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}
	return nil
}
