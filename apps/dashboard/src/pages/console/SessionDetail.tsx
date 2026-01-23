import { Card, List, Tag, Typography, Skeleton } from 'antd'
import { useParams } from 'react-router-dom'
import { PageHeader } from '../../components/PageHeader'
import { RequestBanner } from '../../components/RequestBanner'
import { useRequest } from '../../hooks/useRequest'
import { consoleApi } from '../../services/console'

export function SessionDetail() {
  const { id } = useParams()
  const sessionId = id ?? ''
  const { data, loading, error } = useRequest(
    () => consoleApi.listMessages(sessionId),
    { items: [] },
    { enabled: Boolean(sessionId) },
  )

  return (
    <div className="page">
      <PageHeader title="会话详情" description="消息记录与引用" />
      <RequestBanner error={error} />
      <Card>
        <Typography.Text className="muted">Session ID: {sessionId}</Typography.Text>
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
                      {item.id}
                    </span>
                  }
                  description={
                    <div>
                      <Typography.Paragraph style={{ marginBottom: 4 }}>{item.content}</Typography.Paragraph>
                      <div>
                        <Typography.Text type="secondary">{item.created_at}</Typography.Text>
                      </div>
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
