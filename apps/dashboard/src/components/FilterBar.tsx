import type { ReactNode } from 'react'
import { Space } from 'antd'

type FilterBarProps = {
  left?: ReactNode
  right?: ReactNode
}

export function FilterBar({ left, right }: FilterBarProps) {
  return (
    <div style={{ display: 'flex', justifyContent: 'space-between', gap: 16, flexWrap: 'wrap' }}>
      <Space wrap>{left}</Space>
      <Space wrap>{right}</Space>
    </div>
  )
}
