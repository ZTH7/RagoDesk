import { Card, Collapse, List, Tag, Typography, Skeleton, Table } from 'antd'
import { useParams } from 'react-router-dom'
import { PageHeader } from '../../components/PageHeader'
import { TechnicalMeta } from '../../components/TechnicalMeta'
import { RequestBanner } from '../../components/RequestBanner'
import { useRequest } from '../../hooks/useRequest'
import { consoleApi } from '../../services/console'

export function SessionDetail() {
  const { id } = useParams()
  const sessionId = id ?? ''
  const { data, loading, error } = useRequest(
    () => consoleApi.listMessages(sessionId),
    { items: [] },
    { enabled: Boolean(sessionId), deps: [sessionId] },
  )

  return (
    <div className="page">
      <PageHeader title="会话详情" description="消息记录与引用" />
      <RequestBanner error={error} />
      <Card>
        <TechnicalMeta
          items={[
            { key: 'session-id', label: 'Session ID', value: sessionId },
          ]}
        />
      </Card>
      <Card title="消息记录">
        {loading ? (
          <Skeleton active paragraph={{ rows: 4 }} />
        ) : (
          <List
            dataSource={data.items}
            renderItem={(item) => (
              <List.Item>
                <List.Item.Meta
                  title={
                    <span>
                      <Tag color={item.role === 'assistant' ? 'blue' : 'default'}>{item.role}</Tag>
                      <Typography.Text type="secondary" style={{ marginLeft: 8 }}>
                        {item.created_at}
                      </Typography.Text>
                      {typeof item.confidence === 'number' && (
                        <Tag color="geekblue" style={{ marginLeft: 8 }}>
                          confidence: {(item.confidence * 100).toFixed(1)}%
                        </Tag>
                      )}
                    </span>
                  }
                  description={
                    <div>
                      <Typography.Paragraph style={{ marginBottom: 4 }}>{item.content}</Typography.Paragraph>
                      <div>
                        <Typography.Text type="secondary">{item.created_at}</Typography.Text>
                      </div>
                      {item.references && item.references.length > 0 && (
                        <Collapse
                          ghost
                          style={{ marginTop: 8 }}
                          items={[
                            {
                              key: 'refs',
                              label: `引用来源 (${item.references.length})`,
                              children: (
                                <Table
                                  size="small"
                                  pagination={false}
                                  rowKey={(ref) => `${ref.document_id}-${ref.chunk_id}-${ref.rank}`}
                                  dataSource={item.references}
                                  columns={[
                                    { title: 'Doc', dataIndex: 'document_id' },
                                    { title: 'Version', dataIndex: 'document_version_id' },
                                    { title: 'Chunk', dataIndex: 'chunk_id' },
                                    { title: 'Score', dataIndex: 'score', render: (v) => v.toFixed(3) },
                                    { title: 'Rank', dataIndex: 'rank' },
                                    {
                                      title: 'Snippet',
                                      dataIndex: 'snippet',
                                      render: (v) => v || '-',
                                    },
                                  ]}
                                />
                              ),
                            },
                          ]}
                        />
                      )}
                    </div>
                  }
                />
              </List.Item>
            )}
          />
        )}
      </Card>
    </div>
  )
}
