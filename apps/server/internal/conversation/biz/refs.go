package biz

import "encoding/json"

// Reference describes a retrieval reference for citations.
type Reference struct {
	DocumentID        string  `json:"document_id"`
	DocumentVersionID string  `json:"document_version_id"`
	ChunkID           string  `json:"chunk_id"`
	Score             float32 `json:"score"`
	Rank              int32   `json:"rank"`
	Snippet           string  `json:"snippet"`
}

// EncodeReferences marshals references into JSON for storage.
func EncodeReferences(refs []Reference) string {
	if len(refs) == 0 {
		return ""
	}
	raw, err := json.Marshal(refs)
	if err != nil {
		return ""
	}
	return string(raw)
}

// DecodeReferences unmarshals references from JSON storage.
func DecodeReferences(raw string) []Reference {
	if raw == "" {
		return nil
	}
	var refs []Reference
	if err := json.Unmarshal([]byte(raw), &refs); err != nil {
		return nil
	}
	return refs
}
