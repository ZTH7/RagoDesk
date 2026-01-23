import { Card, Table } from 'antd'
import { PageHeader } from '../../components/PageHeader'
import { DataSourceTag } from '../../components/DataSourceTag'
import { RequestBanner } from '../../components/RequestBanner'
import { ChartPlaceholder } from '../../components/ChartPlaceholder'
import { useRequest } from '../../hooks/useRequest'
import { analyticsApi } from '../../services/analytics'

export function AnalyticsLatency() {
  const { data, loading, source, error } = useRequest(() => analyticsApi.getLatency(), { points: [] })

  return (
    <div className="page">
      <PageHeader title="延迟趋势" description="按天聚合的平均与 P95 延迟" extra={<DataSourceTag source={source} />} />
      <RequestBanner error={error} />
      <Card>
        <ChartPlaceholder title="Latency Chart" description="等待统计 API 数据接入" />
      </Card>
      <Card title="明细">
        <Table
          rowKey="date"
          loading={loading}
          dataSource={data.points}
          columns={[
            { title: '日期', dataIndex: 'date' },
            { title: 'Avg Latency', dataIndex: 'avg_latency_ms' },
            { title: 'P95 Latency', dataIndex: 'p95_latency_ms' },
            { title: 'Total Queries', dataIndex: 'total_queries' },
            { title: 'Hit Queries', dataIndex: 'hit_queries' },
          ]}
        />
      </Card>
    </div>
  )
}
