import { Card, Col, Row, Statistic, Table, Typography } from 'antd'
import { PageHeader } from '../../components/PageHeader'
import { DataSourceTag } from '../../components/DataSourceTag'
import { RequestBanner } from '../../components/RequestBanner'
import { useRequest } from '../../hooks/useRequest'
import { analyticsApi } from '../../services/analytics'

export function AnalyticsOverview() {
  const { data, source, error } = useRequest(() => analyticsApi.getOverview(), {
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
  const { data: topData } = useRequest(() => analyticsApi.getTopQuestions(), { items: [] })
  const { data: gapData } = useRequest(() => analyticsApi.getKBGaps(), { items: [] })

  return (
    <div className="page">
      <PageHeader title="统计总览" description="近 7 天核心指标与趋势概览" extra={<DataSourceTag source={source} />} />
      <RequestBanner error={error} />

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
        <Typography.Text className="muted">
          图表与趋势分析将在对接统计 API 后展示。
        </Typography.Text>
      </Card>
    </div>
  )
}
