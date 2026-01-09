package biz

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"time"
	"unicode"

	"github.com/google/uuid"
)

func buildChunks(rawText, docVersionID string, chunkSize, overlap int) []DocChunk {
	chunkSize = clampInt(chunkSize, 1, 8192)
	overlap = clampInt(overlap, 0, chunkSize-1)
	parts := splitByTokens(rawText, chunkSize, overlap)

	out := make([]DocChunk, 0, len(parts))
	for i, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		chunkIndex := int32(i)
		chunkID := deterministicChunkID(docVersionID, chunkIndex)
		createdAt := time.Now()
		chunk := DocChunk{
			ID:          chunkID,
			ChunkIndex:  chunkIndex,
			Content:     part,
			TokenCount:  int32(estimateTokenCount(part)),
			ContentHash: sha256Hex(part),
			Language:    detectLanguage(part),
			CreatedAt:   createdAt,
		}
		out = append(out, chunk)
	}
	return out
}

type tokenSpan struct {
	start int
	end   int
}

func splitByTokens(s string, chunkSize, overlap int) []string {
	runes := []rune(s)
	if len(runes) == 0 {
		return nil
	}
	tokens := tokenizeSpans(runes)
	if len(tokens) == 0 {
		return nil
	}
	if chunkSize <= 0 {
		chunkSize = len(tokens)
	}
	if overlap < 0 {
		overlap = 0
	}
	if overlap >= chunkSize {
		overlap = chunkSize - 1
	}
	step := chunkSize - overlap
	if step <= 0 {
		step = chunkSize
	}
	out := make([]string, 0, (len(tokens)+step-1)/step)
	for start := 0; start < len(tokens); start += step {
		end := start + chunkSize
		if end > len(tokens) {
			end = len(tokens)
		}
		startIdx := tokens[start].start
		endIdx := tokens[end-1].end
		out = append(out, string(runes[startIdx:endIdx]))
		if end == len(tokens) {
			break
		}
	}
	return out
}

func tokenizeSpans(runes []rune) []tokenSpan {
	out := make([]tokenSpan, 0, len(runes)/2+1)
	for i := 0; i < len(runes); {
		r := runes[i]
		switch {
		case isCJK(r):
			out = append(out, tokenSpan{start: i, end: i + 1})
			i++
		case unicode.IsLetter(r) || unicode.IsDigit(r):
			start := i
			i++
			for i < len(runes) {
				if isCJK(runes[i]) || !(unicode.IsLetter(runes[i]) || unicode.IsDigit(runes[i])) {
					break
				}
				i++
			}
			out = append(out, tokenSpan{start: start, end: i})
		default:
			i++
		}
	}
	return out
}

func estimateTokenCount(s string) int {
	return len(tokenizeSpans([]rune(s)))
}

func detectLanguage(s string) string {
	var cjk, latin int
	for _, r := range s {
		switch {
		case isCJK(r):
			cjk++
		case r <= unicode.MaxLatin1 && (unicode.IsLetter(r) || unicode.IsDigit(r)):
			latin++
		}
	}
	if cjk >= 10 && cjk >= latin {
		return "zh"
	}
	if latin >= 10 && latin > cjk {
		return "en"
	}
	return ""
}

func isCJK(r rune) bool {
	// Common CJK Unified Ideographs blocks.
	return (r >= 0x4E00 && r <= 0x9FFF) || (r >= 0x3400 && r <= 0x4DBF) || (r >= 0x20000 && r <= 0x2A6DF)
}

func deterministicChunkID(docVersionID string, chunkIndex int32) string {
	// Deterministic IDs make ingestion idempotent for the same document_version.
	key := fmt.Sprintf("%s:%d", docVersionID, chunkIndex)
	return uuid.NewSHA1(uuid.NameSpaceOID, []byte(key)).String()
}

func sha256Hex(s string) string {
	sum := sha256.Sum256([]byte(s))
	return hex.EncodeToString(sum[:])
}

func clampInt(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}
