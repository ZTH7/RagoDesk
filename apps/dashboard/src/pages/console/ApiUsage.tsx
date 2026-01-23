import { Input, Select, Tag, Button, Space } from 'antd'
import { useMemo, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { PageHeader } from '../../components/PageHeader'
import { FilterBar } from '../../components/FilterBar'
import { TableCard } from '../../components/TableCard'
import { DataSourceTag } from '../../components/DataSourceTag'
import { RequestBanner } from '../../components/RequestBanner'
import { useRequest } from '../../hooks/useRequest'
import { consoleApi } from '../../services/console'

export function ApiUsage() {
  const [status, setStatus] = useState<string>('all')
  const [keyword, setKeyword] = useState('')
  const navigate = useNavigate()
  const { data, loading, source, error } = useRequest(() => consoleApi.listUsageLogs(), { items: [] })

  const filtered = useMemo(() => {
    return data.items.filter((item) => {
      if (status !== 'all' && String(item.status_code) !== status) return false
      if (keyword && !item.path.toLowerCase().includes(keyword.toLowerCase())) return false
      return true
    })
  }, [data.items, keyword, status])

  return (
    <div className="page">
      <PageHeader
        title="调用日志"
        description="API 使用明细与状态码"
        extra={
          <Space>
            <DataSourceTag source={source} />
            <Button type="default" onClick={() => navigate('/console/api-usage/summary')}>
              查看汇总
            </Button>
          </Space>
        }
      />
      <RequestBanner error={error} />
      <FilterBar
        left={<Input.Search placeholder="搜索路径" onSearch={setKeyword} allowClear style={{ width: 220 }} />}
        right={
          <Select
            value={status}
            style={{ width: 160 }}
            onChange={setStatus}
            options={[
              { value: 'all', label: '全部状态' },
              { value: '200', label: '200' },
              { value: '429', label: '429' },
              { value: '500', label: '500' },
            ]}
          />
        }
      />
      <TableCard
        table={{
          rowKey: 'id',
          dataSource: filtered,
          loading,
          pagination: { pageSize: 10 },
          columns: [
            { title: 'Path', dataIndex: 'path' },
            {
              title: 'Status',
              dataIndex: 'status_code',
              render: (status: number) => <Tag color={status >= 400 ? 'red' : 'green'}>{status}</Tag>,
            },
            { title: 'Latency (ms)', dataIndex: 'latency_ms' },
            { title: 'Total Tokens', dataIndex: 'total_tokens' },
            { title: 'Client IP', dataIndex: 'client_ip' },
            { title: 'Created At', dataIndex: 'created_at' },
          ],
        }}
      />
    </div>
  )
}
