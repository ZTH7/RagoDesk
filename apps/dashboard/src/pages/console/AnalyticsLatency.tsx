import { Button, Card, Col, DatePicker, Row, Select, Space, Statistic, Table, Tag } from 'antd'
import type { Dayjs } from 'dayjs'
import { useMemo, useState } from 'react'
import { PageHeader } from '../../components/PageHeader'
import { DataSourceTag } from '../../components/DataSourceTag'
import { RequestBanner } from '../../components/RequestBanner'
import { FilterBar } from '../../components/FilterBar'
import { useRequest } from '../../hooks/useRequest'
import { consoleApi } from '../../services/console'
import { analyticsApi } from '../../services/analytics'
import { TrendLine } from '../../components/TrendLine'

export function AnalyticsLatency() {
  const [botId, setBotId] = useState<string>('all')
  const [range, setRange] = useState<[Dayjs, Dayjs] | null>(null)
  const [query, setQuery] = useState<{ bot_id?: string; start_time?: string; end_time?: string }>({})

  const { data: botsData } = useRequest(() => consoleApi.listBots(), { items: [] })
  const { data, loading, source, error } = useRequest(() => analyticsApi.getLatency(query), { points: [] })

  const applyFilters = () => {
    const next: { bot_id?: string; start_time?: string; end_time?: string } = {}
    if (botId && botId !== 'all') next.bot_id = botId
    if (range) {
      next.start_time = range[0].toISOString()
      next.end_time = range[1].toISOString()
    }
    setQuery(next)
  }

  const resetFilters = () => {
    setBotId('all')
    setRange(null)
    setQuery({})
  }

  const botOptions = useMemo(
    () => [{ label: '全部 Bot', value: 'all' }].concat(botsData.items.map((bot) => ({ label: bot.name, value: bot.id }))),
    [botsData.items],
  )
  const summary = useMemo(() => {
    if (!data.points.length) {
      return {
        avgLatency: 0,
        p95Latency: 0,
        totalQueries: 0,
        hitRate: 0,
      }
    }
    let avgLatency = 0
    let p95Latency = 0
    let totalQueries = 0
    let hitQueries = 0
    data.points.forEach((point) => {
      avgLatency += point.avg_latency_ms
      p95Latency += point.p95_latency_ms
      totalQueries += point.total_queries
      hitQueries += point.hit_queries
    })
    const count = data.points.length
    return {
      avgLatency: Math.round(avgLatency / count),
      p95Latency: Math.round(p95Latency / count),
      totalQueries,
      hitRate: totalQueries ? hitQueries / totalQueries : 0,
    }
  }, [data.points])

  return (
    <div className="page">
      <PageHeader title="延迟趋势" description="按天聚合的平均与 P95 延迟" extra={<DataSourceTag source={source} />} />
      <RequestBanner error={error} />
      <FilterBar
        left={
          <Space>
            <Select value={botId} onChange={setBotId} options={botOptions} style={{ width: 200 }} />
            <DatePicker.RangePicker value={range} onChange={(value) => setRange(value)} />
          </Space>
        }
        right={
          <Space>
            <Button onClick={resetFilters}>重置</Button>
            <Button type="primary" onClick={applyFilters}>
              应用筛选
            </Button>
          </Space>
        }
      />
      <Row gutter={16}>
        <Col xs={24} md={6}>
          <Card>
            <Statistic title="平均延迟" value={summary.avgLatency} suffix="ms" />
          </Card>
        </Col>
        <Col xs={24} md={6}>
          <Card>
            <Statistic title="P95 延迟" value={summary.p95Latency} suffix="ms" />
          </Card>
        </Col>
        <Col xs={24} md={6}>
          <Card>
            <Statistic title="总请求数" value={summary.totalQueries} />
          </Card>
        </Col>
        <Col xs={24} md={6}>
          <Card>
            <Statistic title="检索命中率" value={summary.hitRate * 100} precision={2} suffix="%" />
          </Card>
        </Col>
      </Row>
      <Card title="趋势图">
        <Space align="center" style={{ marginBottom: 12 }}>
          <Tag color="blue">Avg Latency</Tag>
          <Tag color="purple">P95 Latency</Tag>
        </Space>
        <TrendLine
          series={[
            {
              name: 'avg',
              values: data.points.map((item) => item.avg_latency_ms),
              color: '#1B4B66',
            },
            {
              name: 'p95',
              values: data.points.map((item) => item.p95_latency_ms),
              color: '#6D28D9',
            },
          ]}
        />
      </Card>
      <Card title="明细">
        <Table
          rowKey="date"
          loading={loading}
          dataSource={data.points}
          columns={[
            { title: '日期', dataIndex: 'date' },
            { title: 'Avg Latency', dataIndex: 'avg_latency_ms' },
            { title: 'P95 Latency', dataIndex: 'p95_latency_ms' },
            { title: 'Total Queries', dataIndex: 'total_queries' },
            { title: 'Hit Queries', dataIndex: 'hit_queries' },
          ]}
        />
      </Card>
    </div>
  )
}
