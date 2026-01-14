package biz

import (
	"fmt"
	"strings"
)

const (
	maxSnippetChars = 1200
)

func buildPrompt(question string, ranked []scoredChunk, chunks map[string]ChunkMeta) string {
	var builder strings.Builder
	builder.WriteString("Use the context to answer the question. If the context does not contain the answer, say you don't know.\n\n")
	builder.WriteString("Context:\n")
	count := 0
	for _, item := range ranked {
		meta, ok := chunks[item.result.ChunkID]
		if !ok {
			continue
		}
		count++
		builder.WriteString(formatContextBlock(count, meta))
		builder.WriteString("\n")
	}
	if count == 0 {
		builder.WriteString("(no context available)\n")
	}
	builder.WriteString("\nQuestion: ")
	builder.WriteString(strings.TrimSpace(question))
	builder.WriteString("\nAnswer:")
	return builder.String()
}

func formatContextBlock(index int, meta ChunkMeta) string {
	header := fmt.Sprintf("[%d] doc=%s chunk=%s", index, meta.DocumentID, meta.ChunkID)
	if meta.Section != "" {
		header += " section=" + meta.Section
	}
	if meta.PageNo > 0 {
		header += fmt.Sprintf(" page=%d", meta.PageNo)
	}
	content := truncateText(meta.Content, maxSnippetChars)
	return header + "\n" + content
}

func truncateText(text string, limit int) string {
	text = strings.TrimSpace(text)
	if limit <= 0 || len(text) <= limit {
		return text
	}
	return text[:limit] + "..."
}
