package nexus

import (
	"errors"
	"fmt"
	"io"
	"reflect"
	"strings"

	"github.com/hanjos/nexus"
	"github.com/hanjos/nexus/credentials"
	"github.com/hanjos/nexus/search"
	"github.com/thomasf/nexus-cli/pkg/artifact"
)

// Client .
type Client struct {
	nexus.Client
	NexusURL    string
	Credentials credentials.Credentials
}

func (c *Client) Search(q string) ([]artifact.Artifact, error) {
	var results []artifact.Artifact
	n := nexus.New(c.NexusURL, c.Credentials)

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
		art, err := c.newArtifact(a)
		if err != nil {
			return (make([]artifact.Artifact, 0)), err
		}
		results = append(results, art)
	}
	return results, nil
}

func (c *Client) Fetch(a artifact.Artifact) (io.Reader, error) {
	// artifacts, err := c.Search(a.String())
	na := search.Criteria{
		GroupID:      a.GroupID,
		ArtifactID:   a.ArtifactID,
		Version:      a.Version,
		Classifier:   a.Classifier,
		Extension:    a.Extension,
		RepositoryID: a.RepositoryID,
	}

	n := nexus.New(c.NexusURL, c.Credentials)

	var results []artifact.Artifact

	return nil, errors.New("not implemented")

	arts, err := c.Artifacts(na)
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

func (c *Client) newArtifact(a *nexus.Artifact) (artifact.Artifact, error) {
	// v, _ := version.NewVersion(a.Version)
	return artifact.Artifact{
		Repository:   c,
		GroupID:      a.GroupID,
		ArtifactID:   a.ArtifactID,
		Version:      a.Version,
		Classifier:   a.Classifier,
		Extension:    a.Extension,
		RepositoryID: a.RepositoryID,
	}, nil
}
