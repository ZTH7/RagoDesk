import { Typography } from 'antd'

type TrendSeries = {
  name: string
  values: number[]
  color: string
}

type TrendLineProps = {
  series: TrendSeries[]
  height?: number
}

export function TrendLine({ series, height = 56 }: TrendLineProps) {
  const allValues = series.flatMap((item) => item.values)
  if (allValues.length === 0) {
    return <Typography.Text className="muted">暂无趋势数据</Typography.Text>
  }

  const min = Math.min(...allValues)
  const max = Math.max(...allValues)
  const range = max - min || 1
  const width = 120
  const padding = 6

  const buildPath = (values: number[]) => {
    const safeValues = values.length > 0 ? values : [0]
    const step = safeValues.length > 1 ? (width - padding * 2) / (safeValues.length - 1) : 0
    return safeValues
      .map((value, index) => {
        const x = padding + index * step
        const y = height - padding - ((value - min) / range) * (height - padding * 2)
        return `${index === 0 ? 'M' : 'L'}${x},${y}`
      })
      .join(' ')
  }

  return (
    <svg width="100%" height={height} viewBox={`0 0 ${width} ${height}`} preserveAspectRatio="none">
      {series.map((item) => (
        <path
          key={item.name}
          d={buildPath(item.values)}
          fill="none"
          stroke={item.color}
          strokeWidth="2"
        />
      ))}
    </svg>
  )
}
