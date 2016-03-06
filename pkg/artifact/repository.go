package artifact

import "io"

type Repository interface {
	Fetcher
	// PreRelease
	Searcher
}

// Fetcher fetches an artifact from a repository
type Fetcher interface {
	Fetch(artifact Artifact) (io.Reader, error)
}

// Searcher
type Searcher interface {
	Search(keyword string) ([]Artifact, error)
}

type PreRelease interface {
	PreRelease(Artifact) bool
}
