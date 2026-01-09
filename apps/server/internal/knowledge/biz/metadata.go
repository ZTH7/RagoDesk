package biz

// DocumentMeta describes document-level metadata extracted during parsing.
type DocumentMeta struct {
	Title      string
	SourceURI  string
	SourceType string
}

// DocumentBlock is a structured block of document content.
type DocumentBlock struct {
	Text    string
	Section string
	PageNo  int32
}

// ParsedDocument represents parsed document content and metadata.
type ParsedDocument struct {
	Meta   DocumentMeta
	Blocks []DocumentBlock
}
