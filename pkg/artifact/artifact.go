package artifact

import (
	"errors"
	"sort"
	"strings"

	version "github.com/hashicorp/go-version"
)

// Artifact is the central datastructure, it's default fields are based on the
// maven coordinates structure plus support for semantic versioning.
// All content repository types might not support all these fields.
type Artifact struct {
	GroupID    string // e.g. org.springframework
	ArtifactID string // e.g. spring-core
	Version    string // e.g. 4.1.3.RELEASE
	Classifier string // e.g. sources, javadoc, <the empty string>...
	Extension  string // e.g. jar

	Repository          // related repository
	RepositoryID string // e.g. releases

	semanticVersion *version.Version // semver representation of the Version field.
	// invalidSemver   string           // set to the last parsed failed semver
	// Tags map[string][]string //   additional tags with value support
}

// String implements the fmt.Stringer interface, as per Maven docs
// (http://maven.apache.org/pom.html#Maven_Coordinates).
func (a Artifact) String() string {
	var parts = []string{a.GroupID, a.ArtifactID, a.Extension}

	if a.Classifier != "" {
		parts = append(parts, a.Classifier)
	}

	return strings.Join(append(parts, a.Version), ":") + "@" + a.RepositoryID
}

func Parse(s string) (Artifact, error) {
	var RepositoryID string
	if pos := strings.LastIndex(s, "@"); pos != -1 {
		RepositoryID = s[pos+1:]
		s = s[:pos]
	}
	parts := strings.Split(s, ":")
	var artifact Artifact
	// this is stupid, fix it later
	switch len(parts) {
	case 3:
		artifact = Artifact{
			GroupID:    parts[0],
			ArtifactID: parts[1],
			Version:    parts[2],
			RepositoryID: RepositoryID,
		}
	case 4:
		artifact = Artifact{
			GroupID:    parts[0],
			ArtifactID: parts[1],
			Extension:  parts[2],
			Version:    parts[3],
			RepositoryID: RepositoryID,
		}
	case 5:
		artifact = Artifact{
			GroupID:    parts[0],
			ArtifactID: parts[1],
			Extension:  parts[2],
			Classifier: parts[3],
			Version:    parts[4],
			RepositoryID: RepositoryID,
		}
	default:
		return Artifact{}, errors.New("invalid string")
	}

	return artifact, nil
}

// SorBySemver sorts artifacts acording to their Semantic version number
func SortBySemver(artifacts []Artifact) {
	sort.Sort(BySemver(artifacts))
}

// BySemver is a Sort.sort sorter
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

// GreaterThan compares the semanticVersion version number
func (a *Artifact) GreaterThan(b Artifact) bool {
	return a.semanticVersion.GreaterThan(b.semanticVersion)
}

// LessThan compares the semanticVersion version number
func (a *Artifact) LessThan(b Artifact) bool {
	return a.semanticVersion.LessThan(b.semanticVersion)
}

// Equal compares the semanticVersion version number, nothing else.
func (a *Artifact) Equal(b Artifact) bool {
	return a.semanticVersion.Equal(b.semanticVersion)
}
