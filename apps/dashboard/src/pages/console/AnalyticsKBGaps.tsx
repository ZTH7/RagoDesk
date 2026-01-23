import { Card, Table } from 'antd'
import { PageHeader } from '../../components/PageHeader'
import { DataSourceTag } from '../../components/DataSourceTag'
import { RequestBanner } from '../../components/RequestBanner'
import { useRequest } from '../../hooks/useRequest'
import { analyticsApi } from '../../services/analytics'

export function AnalyticsKBGaps() {
  const { data, loading, source, error } = useRequest(() => analyticsApi.getKBGaps(), { items: [] })

  return (
    <div className="page">
      <PageHeader title="知识缺口" description="低命中 query 与置信度" extra={<DataSourceTag source={source} />} />
      <RequestBanner error={error} />
      <Card>
        <Table
          rowKey="query"
          loading={loading}
          dataSource={data.items}
          columns={[
            { title: 'Query', dataIndex: 'query' },
            { title: 'Miss Count', dataIndex: 'miss_count' },
            { title: 'Avg Confidence', dataIndex: 'avg_confidence', render: (v) => v.toFixed(2) },
            { title: 'Last Seen', dataIndex: 'last_seen_at' },
          ]}
        />
      </Card>
    </div>
  )
}
