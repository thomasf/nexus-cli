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
	"strings"
	// "github.com/hanjos/nexus"
	"github.com/hanjos/nexus/credentials"
	"github.com/hanjos/nexus/search"
	// "github.com/hanjos/nexus/search"
	flags "github.com/jessevdk/go-flags"
	"github.com/thomasf/lg"
	"github.com/thomasf/nexus-cli/pkg/artifact"
	"github.com/thomasf/nexus-cli/pkg/repos/nexus"
)

type Options struct {
	User     string `long:"user" description:"username" ini-name:"username"`
	Password string `long:"password" description:"password" ini-name:"password"`
	Host     string `long:"host" description:"nexus url" ini-name:"host"`
}

type FilterOptions struct {
	Snapshot bool `long:"snapshot" description:"Only show snapshots"`
	Release  bool `long:"release" description:"Only show releases" `
	Latest   bool `long:"latest" description:"Only show latest (semver) versions, implies --release"`
	POM      bool `long:"pom" description:"Show pom results"`
}

func (f *FilterOptions) Filter(artifacts []artifact.Artifact) []artifact.Artifact {
	if !f.POM {
		artifacts = filterPOM(artifacts)
	}
	if f.Latest {
		artifacts = getLatest(artifacts)
		artifact.SortBySemver(artifacts)
	}
	if f.Release {
		artifacts = getReleases(artifacts)
		artifact.SortBySemver(artifacts)

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

func getReleases(artifacts []artifact.Artifact) []artifact.Artifact {
	var result []artifact.Artifact

	for _, v := range artifacts {
		if !strings.HasSuffix(v.Version, "SNAPSHOT") {
			// if v.v == nil {
				// lg.Warningln("could not parse version from", v)
				// continue
			// }

			result = append(result, v)
		}
	}
	return result
}

func getSnapshots(artifacts []artifact.Artifact) []artifact.Artifact {
	var result []artifact.Artifact
	for _, v := range artifacts {
		if strings.HasSuffix(v.Version, "SNAPSHOT") {
			result = append(result, v)
		}

	}
	return result
}

func getLatest(artifacts []artifact.Artifact) []artifact.Artifact {
	var result []artifact.Artifact

	byArtifact := make(map[string][]artifact.Artifact, 0)
	for _, a := range getReleases(artifacts) {
		if !strings.HasSuffix(a.Version, "SNAPSHOT") {
			parts := []string{a.GroupID, a.ArtifactID, a.Extension, a.Classifier, a.RepositoryID}
			key := strings.Join(parts, "\u0000")
			byArtifact[key] = append(byArtifact[key], a)
		}
	}
	for _, v := range byArtifact {
		artifact.SortBySemver(v)
		result = append(result, v[len(v)-1])
	}
	return result
}

func filterPOM(artifacts []artifact.Artifact) []artifact.Artifact {
	var result []artifact.Artifact
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
		return nil
	}
	client := nexus.Client{
		NexusURL:    options.Host,
		Credentials: credentials.BasicAuth(options.User, options.Password),
	}
	artifacts, err := client.Search(args[0])
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
	return errors.New("not done")

	// gav := args[0]

	// client := nexus.Client{
	// 	NexusURL:    options.Host,
	// 	Credentials: credentials.BasicAuth(options.User, options.Password),
	// }
	// artifacts, err := client.Search(args[0])
	// if err != nil {
	// 	return err
	// }

	// arts, err := n.Artifacts(ParseGAV(gav))
	// if err != nil {
	// 	lg.Fatal(err)
	// }
	// var artifacts []Artifact
	// for _, a := range arts {
	// 	art, err := newArtifact(a)
	// 	if err != nil {
	// 		return err
	// 	}
	// 	artifacts = append(artifacts, art)
	// }
	// artifacts = s.FilterOptions.Filter(artifacts)

	// lg.Infoln(artifacts)
	// if len(artifacts) < 1 {
	// 	return errors.New("no matching artifact")
	// }

	// if s.Output != "" {
	// 	if len(artifacts) != 1 {
	// 		return errors.New("cannot use --out with multiple matches")
	// 	}
	// 	v := artifacts[0]
	// 	info, err := n.InfoOf(v.Artifact)
	// 	if err != nil {
	// 		lg.Fatal(err)
	// 	}
	// 	lg.Infoln(info.URL)
	// 	err = download(s.Output, info.URL, creds)
	// 	if err != nil {
	// 		lg.Fatal(err)
	// 	}
	// 	return nil
	// }

	// for _, v := range artifacts {
	// 	info, err := n.InfoOf(v.Artifact)
	// 	if err != nil {
	// 		lg.Fatal(err)
	// 	}
	// 	lg.Infoln(info.URL)

	// 	dirname := filepath.Join(v.GroupID, v.ArtifactID, v.Version)
	// 	filename := filepath.Base(info.URL)
	// 	err = os.MkdirAll(dirname, 0775)
	// 	if err != nil {
	// 		lg.Fatal(err)
	// 	}

	// 	err = download(filepath.Join(dirname, filename), info.URL, creds)
	// 	if err != nil {
	// 		lg.Fatal(err)	// 	}

	// }

	// return nil

}

var getCommand GetCommand

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
