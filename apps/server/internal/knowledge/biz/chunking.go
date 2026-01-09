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

func buildChunksFromBlocks(blocks []DocumentBlock, meta DocumentMeta, docVersionID string, chunkSize, overlap int) []DocChunk {
	chunkSize = clampInt(chunkSize, 64, 8192)
	overlap = clampInt(overlap, 0, chunkSize-1)

	out := make([]DocChunk, 0)
	var index int32
	state := chunkState{}

	for _, block := range blocks {
		rawText := strings.TrimSpace(block.Text)
		if rawText == "" {
			continue
		}
		if isStructureBoundary(state, block) {
			index = flushChunk(&out, &state, meta, docVersionID, index)
		}

		segments := splitBlockSegments(rawText, chunkSize)
		for _, segment := range segments {
			segment = strings.TrimSpace(segment)
			if segment == "" {
				continue
			}
			segmentTokens := estimateTokenCount(segment)
			if state.tokens > 0 && state.tokens+segmentTokens > chunkSize {
				content, section, pageNo, nextIndex := flushChunkWithContent(&out, &state, meta, docVersionID, index)
				index = nextIndex
				if overlap > 0 {
					overlapText := tailByTokens(content, overlap)
					if overlapText != "" && estimateTokenCount(overlapText)+segmentTokens <= chunkSize {
						state.section = section
						state.pageNo = pageNo
						state.addPart(overlapText)
					}
				}
			}
			if state.tokens == 0 {
				state.section = block.Section
				state.pageNo = block.PageNo
			}
			state.addPart(segment)
		}
	}
	if state.tokens > 0 {
		index = flushChunk(&out, &state, meta, docVersionID, index)
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

func splitBlockSegments(text string, maxTokens int) []string {
	maxTokens = clampInt(maxTokens, 64, 8192)
	sentences := splitBySentences(text)
	if len(sentences) == 0 {
		return nil
	}
	segments := make([]string, 0)
	buf := make([]string, 0, 8)
	tokenCount := 0
	flush := func() {
		if len(buf) == 0 {
			return
		}
		segments = append(segments, strings.TrimSpace(strings.Join(buf, " ")))
		buf = buf[:0]
		tokenCount = 0
	}
	for _, sentence := range sentences {
		sentence = strings.TrimSpace(sentence)
		if sentence == "" {
			continue
		}
		tokens := estimateTokenCount(sentence)
		if tokens > maxTokens {
			flush()
			parts := splitByTokens(sentence, maxTokens, 0)
			for _, part := range parts {
				part = strings.TrimSpace(part)
				if part != "" {
					segments = append(segments, part)
				}
			}
			continue
		}
		if tokenCount+tokens > maxTokens && len(buf) > 0 {
			flush()
		}
		buf = append(buf, sentence)
		tokenCount += tokens
	}
	flush()
	return segments
}

func splitBySentences(text string) []string {
	var out []string
	var buf strings.Builder
	flush := func() {
		segment := strings.TrimSpace(buf.String())
		if segment != "" {
			out = append(out, segment)
		}
		buf.Reset()
	}
	for _, r := range text {
		buf.WriteRune(r)
		if isSentenceBoundary(r) {
			flush()
		}
	}
	flush()
	return out
}

func isSentenceBoundary(r rune) bool {
	switch r {
	case '.', '!', '?', ';', '。', '！', '？', '；', '\n', '\r':
		return true
	default:
		return false
	}
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

func tailByTokens(s string, tokens int) string {
	if tokens <= 0 || s == "" {
		return ""
	}
	runes := []rune(s)
	spans := tokenizeSpans(runes)
	if len(spans) == 0 {
		return ""
	}
	if tokens >= len(spans) {
		return strings.TrimSpace(s)
	}
	start := spans[len(spans)-tokens].start
	return strings.TrimSpace(string(runes[start:]))
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

type chunkState struct {
	parts   []string
	tokens  int
	section string
	pageNo  int32
}

func (s *chunkState) addPart(text string) {
	text = strings.TrimSpace(text)
	if text == "" {
		return
	}
	if s.parts == nil {
		s.parts = make([]string, 0, 4)
	}
	s.parts = append(s.parts, text)
	s.tokens += estimateTokenCount(text)
}

func (s *chunkState) reset() {
	s.parts = nil
	s.tokens = 0
	s.section = ""
	s.pageNo = 0
}

func isStructureBoundary(state chunkState, block DocumentBlock) bool {
	if state.tokens == 0 {
		return false
	}
	if block.PageNo != 0 {
		if state.pageNo == 0 || state.pageNo != block.PageNo {
			return true
		}
	}
	if block.Section != "" {
		if state.section == "" || state.section != block.Section {
			return true
		}
	}
	return false
}

func flushChunk(out *[]DocChunk, state *chunkState, meta DocumentMeta, docVersionID string, index int32) int32 {
	_, _, _, next := flushChunkWithContent(out, state, meta, docVersionID, index)
	return next
}

func flushChunkWithContent(out *[]DocChunk, state *chunkState, meta DocumentMeta, docVersionID string, index int32) (string, string, int32, int32) {
	if state.tokens == 0 {
		return "", "", 0, index
	}
	content := strings.TrimSpace(strings.Join(state.parts, "\n\n"))
	if content == "" {
		state.reset()
		return "", "", 0, index
	}
	createdAt := time.Now()
	chunk := DocChunk{
		ID:          deterministicChunkID(docVersionID, index),
		ChunkIndex:  index,
		Content:     content,
		TokenCount:  int32(estimateTokenCount(content)),
		ContentHash: sha256Hex(content),
		Language:    detectLanguage(content),
		Section:     state.section,
		PageNo:      state.pageNo,
		SourceURI:   meta.SourceURI,
		CreatedAt:   createdAt,
	}
	*out = append(*out, chunk)
	section := state.section
	pageNo := state.pageNo
	state.reset()
	return content, section, pageNo, index + 1
}
