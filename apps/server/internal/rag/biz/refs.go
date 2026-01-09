package biz

// Reference describes a retrieval reference for citations.
type Reference struct {
	DocumentID        string
	DocumentVersionID string
	ChunkID           string
	Score             float32
	Rank              int32
	Snippet           string
}

// References is a list of reference items.
type References []Reference
