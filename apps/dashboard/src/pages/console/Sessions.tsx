import { Input, Select, Tag } from 'antd'
import { useMemo, useState } from 'react'
import { Link } from 'react-router-dom'
import { PageHeader } from '../../components/PageHeader'
import { FilterBar } from '../../components/FilterBar'
import { TableCard } from '../../components/TableCard'
import { DataSourceTag } from '../../components/DataSourceTag'
import { RequestBanner } from '../../components/RequestBanner'
import { useRequest } from '../../hooks/useRequest'
import { consoleApi } from '../../services/console'

export function Sessions() {
  const [status, setStatus] = useState<string>('all')
  const [keyword, setKeyword] = useState('')
  const { data, loading, source, error } = useRequest(() => consoleApi.listSessions(), { items: [] })

  const filtered = useMemo(() => {
    return data.items.filter((item) => {
      if (status !== 'all' && item.status !== status) return false
      if (keyword && !item.id.toLowerCase().includes(keyword.toLowerCase())) return false
      return true
    })
  }, [data.items, keyword, status])

  return (
    <div className="page">
      <PageHeader title="会话管理" description="会话列表与状态" extra={<DataSourceTag source={source} />} />
      <RequestBanner error={error} />
      <FilterBar
        left={<Input.Search placeholder="搜索会话 ID" onSearch={setKeyword} allowClear style={{ width: 220 }} />}
        right={
          <Select
            value={status}
            style={{ width: 160 }}
            onChange={setStatus}
            options={[
              { value: 'all', label: '全部状态' },
              { value: 'bot', label: 'Bot' },
              { value: 'closed', label: 'Closed' },
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
            {
              title: 'Session ID',
              dataIndex: 'id',
              render: (value: string) => <Link to={`/console/sessions/${value}`}>{value}</Link>,
            },
            { title: 'Bot ID', dataIndex: 'bot_id' },
            {
              title: '状态',
              dataIndex: 'status',
              render: (value: string) => <Tag color={value === 'closed' ? 'default' : 'blue'}>{value}</Tag>,
            },
            { title: 'Close Reason', dataIndex: 'close_reason' },
            { title: 'User External ID', dataIndex: 'user_external_id' },
            { title: 'Created At', dataIndex: 'created_at' },
          ],
        }}
      />
    </div>
  )
}
