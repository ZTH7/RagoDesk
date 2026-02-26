import type { ReactNode } from 'react'
import type { TableProps } from 'antd'
import { Card, Empty, Table } from 'antd'

type TableCardProps<T extends object> = {
  title?: ReactNode
  extra?: ReactNode
  table: TableProps<T>
}

export function TableCard<T extends object>({ title, extra, table }: TableCardProps<T>) {
  return (
    <Card className="surface-card motion-enter" title={title} extra={extra} bordered={false}>
      <Table<T>
        {...table}
        locale={table.locale ?? { emptyText: <Empty description="暂无数据" /> }}
      />
    </Card>
  )
}
