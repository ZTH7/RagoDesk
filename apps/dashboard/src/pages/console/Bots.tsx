import { Button, Input, Tag, Tooltip } from 'antd'
import { useMemo, useState } from 'react'
import { Link } from 'react-router-dom'
import { PageHeader } from '../../components/PageHeader'
import { FilterBar } from '../../components/FilterBar'
import { TableCard } from '../../components/TableCard'
import { DataSourceTag } from '../../components/DataSourceTag'
import { RequestBanner } from '../../components/RequestBanner'
import { useRequest } from '../../hooks/useRequest'
import { consoleApi } from '../../services/console'

const statusColors: Record<string, string> = {
  active: 'green',
  disabled: 'red',
}

export function Bots() {
  const [keyword, setKeyword] = useState('')
  const { data, loading, error, source } = useRequest(() => consoleApi.listBots(), { items: [] })

  const filtered = useMemo(() => {
    if (!keyword) return data.items
    return data.items.filter((item) => item.name.toLowerCase().includes(keyword.toLowerCase()))
  }, [data.items, keyword])

  return (
    <div className="page">
      <PageHeader
        title="机器人"
        description="管理 Bot 与默认 RAG 流水线"
        extra={<DataSourceTag source={source} />}
      />
      <RequestBanner error={error} />
      <FilterBar
        left={<Input.Search placeholder="搜索 Bot" onSearch={setKeyword} allowClear style={{ width: 220 }} />}
        right={
          <Tooltip title="Bot CRUD 接口尚未开放">
            <Button type="primary" disabled>
              新建机器人
            </Button>
          </Tooltip>
        }
      />
      <TableCard
        table={{
          rowKey: 'id',
          dataSource: filtered,
          loading,
          pagination: { pageSize: 8 },
          columns: [
            {
              title: 'ID',
              dataIndex: 'id',
              render: (value: string) => <Link to={`/console/bots/${value}`}>{value}</Link>,
            },
            { title: '名称', dataIndex: 'name' },
            {
              title: '状态',
              dataIndex: 'status',
              render: (value: string) => <Tag color={statusColors[value] || 'default'}>{value}</Tag>,
            },
            { title: '创建时间', dataIndex: 'created_at' },
          ],
        }}
      />
    </div>
  )
}
