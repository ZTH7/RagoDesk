import { Typography } from 'antd'
import { useMemo, useState } from 'react'

type TrendSeries = {
  name: string
  values: number[]
  color: string
}

type TrendLineProps = {
  series: TrendSeries[]
  height?: number
}

type Point = {
  x: number
  y: number
  value: number
}

export function TrendLine({ series, height = 96 }: TrendLineProps) {
  const allValues = series.flatMap((item) => item.values)
  const [hoverIndex, setHoverIndex] = useState<number | null>(null)
  const hasData = allValues.length > 0

  const width = 360
  const padding = 12
  const min = hasData ? Math.min(...allValues) : 0
  const max = hasData ? Math.max(...allValues) : 1
  const range = max - min || 1
  const maxCount = hasData ? Math.max(...series.map((item) => item.values.length)) : 0
  const count = Math.max(maxCount, 2)
  const step = (width - padding * 2) / (count - 1)

  const seriesPoints = useMemo(() => {
    return series.map((item) => {
      const points: Point[] = Array.from({ length: count }).map((_, index) => {
        const value = item.values[index] ?? item.values[item.values.length - 1] ?? 0
        const x = padding + index * step
        const y = height - padding - ((value - min) / range) * (height - padding * 2)
        return { x, y, value }
      })
      const path = points
        .map((point, index) => `${index === 0 ? 'M' : 'L'}${point.x},${point.y}`)
        .join(' ')
      return { ...item, points, path }
    })
  }, [series, count, step, height, min, range, padding])

  const activeIndex = hoverIndex == null ? count - 1 : hoverIndex
  const hoverX = padding + activeIndex * step

  const yTicks = Array.from({ length: 4 }).map((_, idx) => {
    const ratio = idx / 3
    const y = padding + ratio * (height - padding * 2)
    return { y }
  })

  if (!hasData) {
    return <Typography.Text className="muted">暂无趋势数据</Typography.Text>
  }

  return (
    <svg
      className="trend-line"
      width="100%"
      height={height}
      viewBox={`0 0 ${width} ${height}`}
      preserveAspectRatio="none"
      onMouseLeave={() => setHoverIndex(null)}
      onMouseMove={(event) => {
        const rect = event.currentTarget.getBoundingClientRect()
        const scaleX = width / Math.max(rect.width, 1)
        const x = (event.clientX - rect.left) * scaleX
        const index = Math.max(0, Math.min(count - 1, Math.round((x - padding) / step)))
        setHoverIndex(index)
      }}
    >
      {yTicks.map((tick) => (
        <line
          key={tick.y}
          x1={padding}
          y1={tick.y}
          x2={width - padding}
          y2={tick.y}
          className="trend-grid"
        />
      ))}

      {seriesPoints.map((item) => (
        <g key={item.name}>
          <path d={item.path} fill="none" stroke={item.color} strokeWidth="2.4" className="trend-path" />
          {item.points.map((point, idx) => (
            <circle
              key={`${item.name}-${idx}`}
              cx={point.x}
              cy={point.y}
              r={idx === activeIndex ? 3.4 : 2.2}
              fill={item.color}
              opacity={idx === activeIndex ? 1 : 0.45}
            />
          ))}
        </g>
      ))}

      <line x1={hoverX} y1={padding} x2={hoverX} y2={height - padding} className="trend-hover-line" />

      <g transform={`translate(${Math.min(hoverX + 8, width - 140)},${padding + 2})`}>
        <rect width="132" height={18 + seriesPoints.length * 14} rx="8" className="trend-tooltip-bg" />
        <text x="8" y="13" className="trend-tooltip-title">
          Index {activeIndex + 1}
        </text>
        {seriesPoints.map((item, idx) => (
          <text key={item.name} x="8" y={27 + idx * 14} className="trend-tooltip-row" fill={item.color}>
            {item.name}: {(item.points[activeIndex]?.value ?? 0).toFixed(2)}
          </text>
        ))}
      </g>
    </svg>
  )
}
