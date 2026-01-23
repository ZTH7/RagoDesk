import { Card, Table } from 'antd'
import { PageHeader } from '../../components/PageHeader'
import { DataSourceTag } from '../../components/DataSourceTag'
import { RequestBanner } from '../../components/RequestBanner'
import { useRequest } from '../../hooks/useRequest'
import { analyticsApi } from '../../services/analytics'

export function AnalyticsTopQuestions() {
  const { data, loading, source, error } = useRequest(() => analyticsApi.getTopQuestions(), { items: [] })

  return (
    <div className="page">
      <PageHeader title="热门问题" description="高频 query 与命中率" extra={<DataSourceTag source={source} />} />
      <RequestBanner error={error} />
      <Card>
        <Table
          rowKey="query"
          loading={loading}
          dataSource={data.items}
          columns={[
            { title: 'Query', dataIndex: 'query' },
            { title: 'Count', dataIndex: 'count' },
            { title: 'Hit Rate', dataIndex: 'hit_rate', render: (v) => `${Math.round(v * 100)}%` },
            { title: 'Last Seen', dataIndex: 'last_seen_at' },
          ]}
        />
      </Card>
    </div>
  )
}
