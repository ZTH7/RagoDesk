import { Button, Card, DatePicker, Select, Space, Table } from 'antd'
import type { Dayjs } from 'dayjs'
import { useMemo, useState } from 'react'
import { PageHeader } from '../../components/PageHeader'
import { DataSourceTag } from '../../components/DataSourceTag'
import { RequestBanner } from '../../components/RequestBanner'
import { FilterBar } from '../../components/FilterBar'
import { useRequest } from '../../hooks/useRequest'
import { consoleApi } from '../../services/console'
import { analyticsApi } from '../../services/analytics'

export function AnalyticsKBGaps() {
  const [botId, setBotId] = useState<string>('all')
  const [range, setRange] = useState<[Dayjs, Dayjs] | null>(null)
  const [query, setQuery] = useState<{ bot_id?: string; start_time?: string; end_time?: string }>({})

  const { data: botsData } = useRequest(() => consoleApi.listBots(), { items: [] })
  const { data, loading, source, error } = useRequest(() => analyticsApi.getKBGaps(query), { items: [] })

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

  return (
    <div className="page">
      <PageHeader title="知识缺口" description="低命中 query 与置信度" extra={<DataSourceTag source={source} />} />
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
      <Card>
        <Table
          rowKey="query"
          loading={loading}
          dataSource={data.items}
          columns={[
            { title: 'Query', dataIndex: 'query' },
            { title: 'Miss Count', dataIndex: 'miss_count' },
            { title: 'Avg Confidence', dataIndex: 'avg_confidence', render: (v) => v.toFixed(2) },
            { title: 'Last Seen', dataIndex: 'last_seen_at' },
          ]}
        />
      </Card>
    </div>
  )
}
