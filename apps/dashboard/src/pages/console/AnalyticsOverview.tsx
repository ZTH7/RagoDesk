import { Button, Card, Col, DatePicker, Row, Select, Space, Statistic, Steps, Table, Tag, Typography } from 'antd'
import type { Dayjs } from 'dayjs'
import { useMemo, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { PageHeader } from '../../components/PageHeader'
import { DataSourceTag } from '../../components/DataSourceTag'
import { RequestBanner } from '../../components/RequestBanner'
import { FilterBar } from '../../components/FilterBar'
import { useRequest } from '../../hooks/useRequest'
import { consoleApi } from '../../services/console'
import { analyticsApi } from '../../services/analytics'
import { TrendLine } from '../../components/TrendLine'

export function AnalyticsOverview() {
  const navigate = useNavigate()
  const [botId, setBotId] = useState<string>('all')
  const [range, setRange] = useState<[Dayjs | null, Dayjs | null] | null>(null)
  const [query, setQuery] = useState<{ bot_id?: string; start_time?: string; end_time?: string }>({})

  const { data: botsData } = useRequest(() => consoleApi.listBots(), { items: [] })
  const { data: kbData } = useRequest(() => consoleApi.listKnowledgeBases(), { items: [] })
  const { data: docData } = useRequest(() => consoleApi.listDocuments({ limit: 200 }), { items: [] })
  const { data: apiKeyData } = useRequest(() => consoleApi.listApiKeys({ limit: 200 }), { items: [] })

  const { data, source, error } = useRequest(
    () => analyticsApi.getOverview(query),
    {
      overview: {
        total_queries: 0,
        hit_queries: 0,
        hit_rate: 0,
        avg_latency_ms: 0,
        p95_latency_ms: 0,
        error_count: 0,
        error_rate: 0,
      },
    },
    { deps: [query] },
  )
  const { data: latencyData } = useRequest(() => analyticsApi.getLatency(query), { points: [] }, { deps: [query] })
  const { data: topData } = useRequest(() => analyticsApi.getTopQuestions(query), { items: [] }, { deps: [query] })
  const { data: gapData } = useRequest(() => analyticsApi.getKBGaps(query), { items: [] }, { deps: [query] })

  const applyFilters = () => {
    const next: { bot_id?: string; start_time?: string; end_time?: string } = {}
    if (botId && botId !== 'all') next.bot_id = botId
    if (range && range[0] && range[1]) {
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

  const readyDocumentCount = useMemo(
    () => docData.items.filter((item) => item.status === 'ready').length,
    [docData.items],
  )
  const hasKb = kbData.items.length > 0
  const hasReadyDoc = readyDocumentCount > 0
  const hasBot = botsData.items.length > 0
  const hasApiKey = apiKeyData.items.length > 0
  const onboardingFinished = hasKb && hasReadyDoc && hasBot && hasApiKey

  return (
    <div className="page">
      <PageHeader title="统计总览" description="近 7 天核心指标与趋势概览" extra={<DataSourceTag source={source} />} />
      <RequestBanner error={error} />
      <Card title="快速上手（推荐流程）" className="surface-card">
        <Typography.Paragraph className="muted" style={{ marginBottom: 12 }}>
          先完成这 4 步，再进入 API 调试与线上接入。
        </Typography.Paragraph>
        <Steps
          responsive
          items={[
            {
              title: '创建知识库',
              description: hasKb ? `已创建 ${kbData.items.length} 个` : '尚未创建',
              status: hasKb ? 'finish' : 'process',
            },
            {
              title: '上传文档并入库',
              description: hasReadyDoc ? `可用文档 ${readyDocumentCount} 个` : '暂无可用文档',
              status: hasReadyDoc ? 'finish' : 'process',
            },
            {
              title: '创建机器人',
              description: hasBot ? `已创建 ${botsData.items.length} 个` : '尚未创建',
              status: hasBot ? 'finish' : 'process',
            },
            {
              title: '创建 API Key',
              description: hasApiKey ? `已创建 ${apiKeyData.items.length} 个` : '尚未创建',
              status: hasApiKey ? 'finish' : 'process',
            },
          ]}
        />
        <Space wrap style={{ marginTop: 14 }}>
          <Button onClick={() => navigate('/console/knowledge-bases')}>去知识库</Button>
          <Button onClick={() => navigate('/console/documents')}>去文档管理</Button>
          <Button onClick={() => navigate('/console/bots')}>去机器人</Button>
          <Button onClick={() => navigate('/console/api-keys')}>去 API Keys</Button>
          <Button type={onboardingFinished ? 'primary' : 'default'} onClick={() => navigate('/console/devtools/api')}>
            去 API 调试
          </Button>
        </Space>
      </Card>

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
          <Card title="延迟趋势">
            <Space align="center" style={{ marginBottom: 12 }}>
              <Tag color="blue">Avg Latency</Tag>
              <Tag color="purple">P95 Latency</Tag>
            </Space>
            <TrendLine
              series={[
                {
                  name: 'avg',
                  values: latencyData.points.map((item) => item.avg_latency_ms),
                  color: '#1B4B66',
                },
                {
                  name: 'p95',
                  values: latencyData.points.map((item) => item.p95_latency_ms),
                  color: '#6D28D9',
                },
              ]}
            />
          </Card>
        </Col>
        <Col xs={24} lg={12}>
          <Card title="命中率趋势">
            <Space align="center" style={{ marginBottom: 12 }}>
              <Tag color="green">Hit Rate</Tag>
            </Space>
            <TrendLine
              series={[
                {
                  name: 'hit_rate',
                  values: latencyData.points.map((item) =>
                    item.total_queries ? item.hit_queries / item.total_queries : 0,
                  ),
                  color: '#16A34A',
                },
              ]}
            />
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
    </div>
  )
}
