import { Card, Col, Row, Statistic, Skeleton } from 'antd'
import { PageHeader } from '../../components/PageHeader'
import { RequestBanner } from '../../components/RequestBanner'
import { useRequest } from '../../hooks/useRequest'
import { consoleApi } from '../../services/console'

export function ApiUsageSummary() {
  const { data, loading, error } = useRequest(() => consoleApi.getUsageSummary({}), {
    summary: {
      total: 0,
      error_count: 0,
      avg_latency_ms: 0,
      prompt_tokens: 0,
      completion_tokens: 0,
      total_tokens: 0,
    },
  })

  return (
    <div className="page">
      <PageHeader title="调用汇总" description="按时间范围汇总的 API 使用情况" />
      <RequestBanner error={error} />
      {loading ? (
        <Skeleton active paragraph={{ rows: 3 }} />
      ) : (
        <Row gutter={16}>
          <Col xs={24} md={6}>
            <Card>
              <Statistic title="总请求" value={data.summary.total} />
            </Card>
          </Col>
          <Col xs={24} md={6}>
            <Card>
              <Statistic title="错误数" value={data.summary.error_count} />
            </Card>
          </Col>
          <Col xs={24} md={6}>
            <Card>
              <Statistic title="平均延迟" value={data.summary.avg_latency_ms} suffix="ms" />
            </Card>
          </Col>
          <Col xs={24} md={6}>
            <Card>
              <Statistic title="总 Token" value={data.summary.total_tokens} />
            </Card>
          </Col>
        </Row>
      )}
    </div>
  )
}
