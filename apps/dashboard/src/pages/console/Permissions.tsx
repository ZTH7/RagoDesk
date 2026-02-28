import { Input, Tag } from 'antd'
import { useMemo, useState } from 'react'
import { PageHeader } from '../../components/PageHeader'
import { FilterBar } from '../../components/FilterBar'
import { TableCard } from '../../components/TableCard'
import { DataSourceTag } from '../../components/DataSourceTag'
import { RequestBanner } from '../../components/RequestBanner'
import { useRequest } from '../../hooks/useRequest'
import { consoleApi } from '../../services/console'

export function Permissions() {
  const [keyword, setKeyword] = useState('')
  const { data, loading, source, error } = useRequest(() => consoleApi.listPermissions(), { items: [] })

  const filtered = useMemo(() => {
    if (!keyword) return data.items
    return data.items.filter((item) => item.code.toLowerCase().includes(keyword.toLowerCase()))
  }, [data.items, keyword])

  return (
    <div className="page">
      <PageHeader title="权限目录" description="租户权限列表" extra={<DataSourceTag source={source} />} />
      <RequestBanner error={error} />
      <FilterBar left={<Input.Search placeholder="搜索权限" onSearch={setKeyword} allowClear style={{ width: 220 }} />} />
      <TableCard
        table={{
          rowKey: 'code',
          dataSource: filtered,
          loading,
          pagination: { pageSize: 10 },
          columns: [
            { title: '权限标识', dataIndex: 'code' },
            { title: '描述', dataIndex: 'description' },
            {
              title: '权限域',
              dataIndex: 'scope',
              render: (scope: string) => <Tag>{scope === 'tenant' ? '租户域' : scope}</Tag>,
            },
          ],
        }}
      />
    </div>
  )
}
