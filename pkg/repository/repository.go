package repository

import (
	"io"

	"github.com/thomasf/nexus-cli/pkg/artifact"
)

type Repository interface {
	Fetcher
	PreRelease
}


// Fetcher fetches an artifact from a repository
type Fetcher interface {
	Fetch(artifact artifact.Artifact) (io.Reader, error)
}

// Searcher
type Searcher interface {
	Search(keyword string) ([]artifact.Artifact, error)
}


type PreRelease interface {
	PreRelease(artifact.Artifact)	bool

}
