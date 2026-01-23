import { Tag } from 'antd'

type DataSourceTagProps = {
  source?: 'api' | 'fallback' | 'empty'
}

export function DataSourceTag({ source }: DataSourceTagProps) {
  if (source === 'api') return <Tag color="green">Live</Tag>
  if (source === 'fallback') return <Tag color="gold">Fallback</Tag>
  return null
}
