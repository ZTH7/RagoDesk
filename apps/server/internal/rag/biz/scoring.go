package biz

import (
	"sort"
	"strings"
	"unicode"
)

type scoredChunk struct {
	result      VectorSearchResult
	vectorScore float32
	textScore   float32
	score       float32
}

func rankAndFilter(items []scoredChunk, topK int) []scoredChunk {
	if len(items) == 0 {
		return nil
	}
	byChunk := make(map[string]scoredChunk, len(items))
	for _, item := range items {
		if item.result.ChunkID == "" {
			continue
		}
		prev, ok := byChunk[item.result.ChunkID]
		if !ok || item.score > prev.score {
			byChunk[item.result.ChunkID] = item
		}
	}
	merged := make([]scoredChunk, 0, len(byChunk))
	for _, item := range byChunk {
		merged = append(merged, item)
	}
	sort.SliceStable(merged, func(i, j int) bool {
		return merged[i].score > merged[j].score
	})
	if topK > 0 && len(merged) > topK {
		merged = merged[:topK]
	}
	return merged
}

func computeConfidence(ranked []scoredChunk, topK int) float32 {
	if len(ranked) == 0 {
		return 0
	}
	limit := 3
	if len(ranked) < limit {
		limit = len(ranked)
	}
	var sum float32
	for i := 0; i < limit; i++ {
		sum += ranked[i].score
	}
	avg := sum / float32(limit)
	coverage := float32(1)
	if topK > 0 {
		coverage = float32(len(ranked)) / float32(topK)
		if coverage > 1 {
			coverage = 1
		}
	}
	conf := 0.8*avg + 0.2*coverage
	if conf < 0 {
		return 0
	}
	if conf > 1 {
		return 1
	}
	return conf
}

func combineScores(vectorScore float32, textScore float32, weight float32) float32 {
	if weight <= 0 {
		return vectorScore
	}
	if weight >= 1 {
		return textScore
	}
	return vectorScore*(1-weight) + textScore*weight
}

func overlapScore(question string, content string) float32 {
	qTokens := tokenSet(question, 64)
	if len(qTokens) == 0 {
		return 0
	}
	cTokens := tokenSet(content, 256)
	if len(cTokens) == 0 {
		return 0
	}
	matched := 0
	for token := range qTokens {
		if _, ok := cTokens[token]; ok {
			matched++
		}
	}
	return float32(matched) / float32(len(qTokens))
}

func tokenSet(text string, limit int) map[string]struct{} {
	if limit <= 0 {
		limit = 64
	}
	fields := strings.FieldsFunc(strings.ToLower(text), func(r rune) bool {
		return !(unicode.IsLetter(r) || unicode.IsNumber(r))
	})
	out := make(map[string]struct{})
	count := 0
	for _, token := range fields {
		if len(token) < 2 {
			continue
		}
		out[token] = struct{}{}
		count++
		if count >= limit {
			break
		}
	}
	return out
}

func maxFloat32(a float32, b float32) float32 {
	if a >= b {
		return a
	}
	return b
}

func deriveRetrieveThreshold(confidenceThreshold float32) float32 {
	if confidenceThreshold <= 0 {
		return 0
	}
	threshold := confidenceThreshold * 0.2
	if threshold > 0.2 {
		threshold = 0.2
	}
	return threshold
}
