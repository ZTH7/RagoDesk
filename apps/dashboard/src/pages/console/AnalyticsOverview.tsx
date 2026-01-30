import { Button, Card, Col, DatePicker, Row, Select, Space, Statistic, Table, Typography } from 'antd'
import type { Dayjs } from 'dayjs'
import { useMemo, useState } from 'react'
import { PageHeader } from '../../components/PageHeader'
import { DataSourceTag } from '../../components/DataSourceTag'
import { RequestBanner } from '../../components/RequestBanner'
import { FilterBar } from '../../components/FilterBar'
import { useRequest } from '../../hooks/useRequest'
import { consoleApi } from '../../services/console'
import { analyticsApi } from '../../services/analytics'

export function AnalyticsOverview() {
  const [botId, setBotId] = useState<string>('all')
  const [range, setRange] = useState<[Dayjs, Dayjs] | null>(null)
  const [query, setQuery] = useState<{ bot_id?: string; start_time?: string; end_time?: string }>({})

  const { data: botsData } = useRequest(() => consoleApi.listBots(), { items: [] })

  const { data, source, error } = useRequest(() => analyticsApi.getOverview(query), {
    overview: {
      total_queries: 0,
      hit_queries: 0,
      hit_rate: 0,
      avg_latency_ms: 0,
      p95_latency_ms: 0,
      error_count: 0,
      error_rate: 0,
    },
  })
  const { data: topData } = useRequest(() => analyticsApi.getTopQuestions(query), { items: [] })
  const { data: gapData } = useRequest(() => analyticsApi.getKBGaps(query), { items: [] })

  const applyFilters = () => {
    const next: { bot_id?: string; start_time?: string; end_time?: string } = {}
    if (botId && botId !== 'all') next.bot_id = botId
    if (range) {
      next.start_time = range[0].toISOString()
      next.end_time = range[1].toISOString()
    }
    setQuery(next)
  }

  const resetFilters = () => {
    setBotId('all')
    setRange(null)
    setQuery({})
  }

  const botOptions = useMemo(
    () => [{ label: '全部 Bot', value: 'all' }].concat(botsData.items.map((bot) => ({ label: bot.name, value: bot.id }))),
    [botsData.items],
  )

  return (
    <div className="page">
      <PageHeader title="统计总览" description="近 7 天核心指标与趋势概览" extra={<DataSourceTag source={source} />} />
      <RequestBanner error={error} />

      <FilterBar
        left={
          <Space>
            <Select value={botId} onChange={setBotId} options={botOptions} style={{ width: 200 }} />
            <DatePicker.RangePicker value={range} onChange={(value) => setRange(value)} />
          </Space>
        }
        right={
          <Space>
            <Button onClick={resetFilters}>重置</Button>
            <Button type="primary" onClick={applyFilters}>
              应用筛选
            </Button>
          </Space>
        }
      />

      <Row gutter={16}>
        <Col xs={24} md={6}>
          <Card>
            <Statistic title="总请求数" value={data.overview.total_queries} />
          </Card>
        </Col>
        <Col xs={24} md={6}>
          <Card>
            <Statistic title="检索命中率" value={data.overview.hit_rate * 100} precision={2} suffix="%" />
          </Card>
        </Col>
        <Col xs={24} md={6}>
          <Card>
            <Statistic title="P95 延迟" value={data.overview.p95_latency_ms} suffix="ms" />
          </Card>
        </Col>
        <Col xs={24} md={6}>
          <Card>
            <Statistic title="错误率" value={data.overview.error_rate * 100} precision={2} suffix="%" />
          </Card>
        </Col>
      </Row>

      <Row gutter={16}>
        <Col xs={24} lg={12}>
          <Card title="热门问题">
            <Table
              size="small"
              pagination={false}
              rowKey="query"
              dataSource={topData.items}
              columns={[
                { title: 'Query', dataIndex: 'query' },
                { title: 'Count', dataIndex: 'count' },
                { title: 'Hit Rate', dataIndex: 'hit_rate', render: (v) => `${Math.round(v * 100)}%` },
                { title: 'Last Seen', dataIndex: 'last_seen_at' },
              ]}
            />
          </Card>
        </Col>
        <Col xs={24} lg={12}>
          <Card title="知识缺口">
            <Table
              size="small"
              pagination={false}
              rowKey="query"
              dataSource={gapData.items}
              columns={[
                { title: 'Query', dataIndex: 'query' },
                { title: 'Miss Count', dataIndex: 'miss_count' },
                { title: 'Avg Confidence', dataIndex: 'avg_confidence', render: (v) => v.toFixed(2) },
                { title: 'Last Seen', dataIndex: 'last_seen_at' },
              ]}
            />
          </Card>
        </Col>
      </Row>

      <Card>
        <Typography.Text className="muted">图表与趋势分析将在对接统计 API 后展示。</Typography.Text>
      </Card>
    </div>
  )
}
