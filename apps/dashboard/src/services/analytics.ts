import { request } from './client'

export type Overview = {
  total_queries: number
  hit_queries: number
  hit_rate: number
  avg_latency_ms: number
  p95_latency_ms: number
  error_count: number
  error_rate: number
}

export type LatencyPoint = {
  date: string
  avg_latency_ms: number
  p95_latency_ms: number
  total_queries: number
  hit_queries: number
}

export type QuestionStat = {
  query: string
  count: number
  hit_rate: number
  last_seen_at: string
}

export type GapStat = {
  query: string
  miss_count: number
  avg_confidence: number
  last_seen_at: string
}

export const analyticsApi = {
  getOverview() {
    return request<{ overview: Overview }>('/console/v1/analytics/overview')
  },
  getLatency() {
    return request<{ points: LatencyPoint[] }>('/console/v1/analytics/latency')
  },
  getTopQuestions() {
    return request<{ items: QuestionStat[] }>('/console/v1/analytics/top_questions')
  },
  getKBGaps() {
    return request<{ items: GapStat[] }>('/console/v1/analytics/kb_gaps')
  },
}
