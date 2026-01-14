package biz

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"unicode"
)

const (
	maxSnippetChars  = 1200
	maxContextBlocks = 12
	maxBlocksPerDoc  = 3
)

func buildPrompt(question string, ranked []scoredChunk, chunks map[string]ChunkMeta) string {
	var builder strings.Builder
	builder.WriteString("Use the context to answer the question. If the context does not contain the answer, say you don't know.\n\n")
	builder.WriteString("Context:\n")
	selected := selectContext(ranked, chunks)
	for idx, meta := range selected {
		builder.WriteString(formatContextBlock(idx+1, meta))
		builder.WriteString("\n")
	}
	if len(selected) == 0 {
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
	content := truncateText(normalizeForPrompt(meta.Content), maxSnippetChars)
	return header + "\n" + content
}

func truncateText(text string, limit int) string {
	text = strings.TrimSpace(text)
	if limit <= 0 || len(text) <= limit {
		return text
	}
	return text[:limit] + "..."
}

func selectContext(ranked []scoredChunk, chunks map[string]ChunkMeta) []ChunkMeta {
	if len(ranked) == 0 || len(chunks) == 0 {
		return nil
	}
	seen := make(map[string]struct{}, len(ranked))
	perDoc := make(map[string]int)
	out := make([]ChunkMeta, 0, len(ranked))
	for _, item := range ranked {
		meta, ok := chunks[item.result.ChunkID]
		if !ok || strings.TrimSpace(meta.Content) == "" {
			continue
		}
		key := contentFingerprint(meta.Content)
		if _, exists := seen[key]; exists {
			continue
		}
		docID := strings.TrimSpace(meta.DocumentID)
		if docID != "" && perDoc[docID] >= maxBlocksPerDoc {
			continue
		}
		seen[key] = struct{}{}
		if docID != "" {
			perDoc[docID]++
		}
		out = append(out, meta)
		if len(out) >= maxContextBlocks {
			break
		}
	}
	return out
}

func contentFingerprint(text string) string {
	text = normalizeForPrompt(text)
	sum := sha256.Sum256([]byte(strings.ToLower(text)))
	return hex.EncodeToString(sum[:])
}

func normalizeForPrompt(text string) string {
	text = strings.TrimSpace(text)
	if text == "" {
		return ""
	}
	var b strings.Builder
	b.Grow(len(text))
	space := false
	for _, r := range text {
		if unicode.IsSpace(r) {
			if !space {
				b.WriteRune(' ')
				space = true
			}
			continue
		}
		b.WriteRune(r)
		space = false
	}
	return b.String()
}
