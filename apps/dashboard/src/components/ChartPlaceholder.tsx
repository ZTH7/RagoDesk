import { Typography } from 'antd'

type ChartPlaceholderProps = {
  title: string
  description?: string
}

export function ChartPlaceholder({ title, description }: ChartPlaceholderProps) {
  return (
    <div className="chart-placeholder">
      <div style={{ textAlign: 'center' }}>
        <Typography.Text strong>{title}</Typography.Text>
        {description ? (
          <div className="muted" style={{ marginTop: 4 }}>
            {description}
          </div>
        ) : null}
      </div>
    </div>
  )
}
