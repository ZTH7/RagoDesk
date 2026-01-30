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

export type AnalyticsQuery = {
  bot_id?: string
  start_time?: string
  end_time?: string
}

function buildQuery(params?: AnalyticsQuery) {
  const query = new URLSearchParams()
  if (params?.bot_id) query.set('bot_id', params.bot_id)
  if (params?.start_time) query.set('start_time', params.start_time)
  if (params?.end_time) query.set('end_time', params.end_time)
  const suffix = query.toString() ? `?${query.toString()}` : ''
  return suffix
}

export const analyticsApi = {
  getOverview(params?: AnalyticsQuery) {
    return request<{ overview: Overview }>(`/console/v1/analytics/overview${buildQuery(params)}`)
  },
  getLatency(params?: AnalyticsQuery) {
    return request<{ points: LatencyPoint[] }>(`/console/v1/analytics/latency${buildQuery(params)}`)
  },
  getTopQuestions(params?: AnalyticsQuery) {
    return request<{ items: QuestionStat[] }>(`/console/v1/analytics/top_questions${buildQuery(params)}`)
  },
  getKBGaps(params?: AnalyticsQuery) {
    return request<{ items: GapStat[] }>(`/console/v1/analytics/kb_gaps${buildQuery(params)}`)
  },
}
