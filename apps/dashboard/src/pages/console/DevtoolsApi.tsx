import {
  Alert,
  Button,
  Card,
  Descriptions,
  Form,
  Input,
  InputNumber,
  Space,
  Table,
  Typography,
  message,
} from 'antd'
import { useState } from 'react'
import { PageHeader } from '../../components/PageHeader'
import { publicApi } from '../../services/public'

export function DevtoolsApi() {
  const [form] = Form.useForm()
  const [sessionId, setSessionId] = useState('')
  const [reply, setReply] = useState('')
  const [confidence, setConfidence] = useState<number | null>(null)
  const [references, setReferences] = useState<any[]>([])
  const [sessionInfo, setSessionInfo] = useState<any | null>(null)
  const [sessionMessages, setSessionMessages] = useState<any[]>([])
  const [loading, setLoading] = useState(false)

  const handleCreateSession = async () => {
    try {
      const values = await form.validateFields(['api_key', 'user_external_id', 'metadata'])
      const apiKey = values.api_key as string
      if (!apiKey) {
        message.error('请先输入 API Key')
        return
      }
      let metadata: Record<string, unknown> | undefined = undefined
      if (values.metadata) {
        metadata = JSON.parse(values.metadata)
      }

      setLoading(true)
      const res = await publicApi.createSession(apiKey, {
        user_external_id: values.user_external_id,
        metadata,
      })
      setSessionId(res.session.id)
      form.setFieldsValue({ session_id: res.session.id })
      message.success('已创建会话')
    } catch (err) {
      if (err instanceof Error) message.error(err.message)
    } finally {
      setLoading(false)
    }
  }

  const handleSend = async () => {
    try {
      const values = await form.validateFields(['api_key', 'session_id', 'message', 'top_k', 'threshold'])
      const apiKey = values.api_key as string
      if (!apiKey) {
        message.error('请先输入 API Key')
        return
      }
      setLoading(true)
      const res = await publicApi.sendMessage(apiKey, {
        session_id: values.session_id,
        message: values.message,
        top_k: values.top_k,
        threshold: values.threshold,
      })
      setReply(res.reply)
      setConfidence(res.confidence)
      setReferences(res.references || [])
      message.success('已收到回复')
    } catch (err) {
      if (err instanceof Error) message.error(err.message)
    } finally {
      setLoading(false)
    }
  }

  const handleGetSession = async () => {
    try {
      const values = await form.validateFields(['api_key', 'session_id'])
      const apiKey = values.api_key as string
      if (!apiKey) {
        message.error('请先输入 API Key')
        return
      }
      setLoading(true)
      const res = await publicApi.getSession(apiKey, values.session_id, { include_messages: true })
      setSessionInfo(res.session)
      setSessionMessages(res.messages || [])
      message.success('已获取会话信息')
    } catch (err) {
      if (err instanceof Error) message.error(err.message)
    } finally {
      setLoading(false)
    }
  }

  const handleCloseSession = async () => {
    try {
      const values = await form.validateFields(['api_key', 'session_id', 'close_reason'])
      const apiKey = values.api_key as string
      if (!apiKey) {
        message.error('请先输入 API Key')
        return
      }
      setLoading(true)
      await publicApi.closeSession(apiKey, values.session_id, values.close_reason)
      message.success('已关闭会话')
    } catch (err) {
      if (err instanceof Error) message.error(err.message)
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="page">
      <PageHeader title="API 调试" description="使用 X-API-Key 调用外部接口" />
      <Card title="快速调试">
        <Form
          form={form}
          layout="vertical"
          initialValues={{
            top_k: 6,
            threshold: 0.2,
          }}
        >
          <Form.Item label="X-API-Key" name="api_key" rules={[{ required: true, message: '请输入 API Key' }]}>
            <Input.Password placeholder="粘贴 API Key" />
          </Form.Item>
          <Form.Item label="User External ID" name="user_external_id">
            <Input placeholder="可选，创建会话时写入" />
          </Form.Item>
          <Form.Item label="Metadata (JSON)" name="metadata">
            <Input.TextArea placeholder='可选，例如：{"source":"web"}' rows={2} />
          </Form.Item>
          <Space>
            <Button onClick={handleCreateSession} loading={loading}>
              创建会话
            </Button>
            {sessionId && <Typography.Text>当前 Session: {sessionId}</Typography.Text>}
          </Space>
          <div style={{ height: 16 }} />
          <Form.Item label="Session ID" name="session_id" rules={[{ required: true, message: '请输入 Session ID' }]}>
            <Input placeholder="填入已有 Session ID" />
          </Form.Item>
          <Form.Item label="Close Reason" name="close_reason">
            <Input placeholder="可选，例如：user_end" />
          </Form.Item>
          <Space>
            <Button onClick={handleGetSession} loading={loading}>
              获取会话
            </Button>
            <Button onClick={handleCloseSession} loading={loading}>
              关闭会话
            </Button>
          </Space>
          <div style={{ height: 16 }} />
          <Form.Item label="Message" name="message" rules={[{ required: true, message: '请输入问题' }]}>
            <Input.TextArea placeholder="请输入问题" rows={3} />
          </Form.Item>
          <Space style={{ display: 'flex' }}>
            <Form.Item label="Top K" name="top_k" style={{ flex: 1 }}>
              <InputNumber min={1} max={50} style={{ width: '100%' }} />
            </Form.Item>
            <Form.Item label="Threshold" name="threshold" style={{ flex: 1 }}>
              <InputNumber min={0} max={1} step={0.05} style={{ width: '100%' }} />
            </Form.Item>
          </Space>
          <Button type="primary" onClick={handleSend} loading={loading}>
            发送消息
          </Button>
        </Form>
      </Card>

      <Card title="响应结果">
        {reply ? (
          <>
            <Typography.Paragraph>{reply}</Typography.Paragraph>
            <Alert
              type="info"
              message={`Confidence: ${confidence ?? 0}`}
              showIcon
              style={{ marginBottom: 16 }}
            />
            <Table
              rowKey={(record) => `${record.document_id}-${record.chunk_id}-${record.rank}`}
              dataSource={references}
              pagination={false}
              columns={[
                { title: 'Doc ID', dataIndex: 'document_id' },
                { title: 'Chunk ID', dataIndex: 'chunk_id' },
                { title: 'Score', dataIndex: 'score' },
                { title: 'Rank', dataIndex: 'rank' },
                { title: 'Snippet', dataIndex: 'snippet' },
              ]}
            />
          </>
        ) : (
          <Alert type="info" message="暂无响应，请发送请求" showIcon />
        )}
      </Card>

      <Card title="会话信息">
        {sessionInfo ? (
          <>
            <Descriptions column={1} bordered size="middle">
              <Descriptions.Item label="Session ID">{sessionInfo.id}</Descriptions.Item>
              <Descriptions.Item label="Status">{sessionInfo.status}</Descriptions.Item>
              <Descriptions.Item label="Bot ID">{sessionInfo.bot_id || '-'}</Descriptions.Item>
              <Descriptions.Item label="User External ID">{sessionInfo.user_external_id || '-'}</Descriptions.Item>
              <Descriptions.Item label="Created At">{sessionInfo.created_at || '-'}</Descriptions.Item>
            </Descriptions>
            <Table
              style={{ marginTop: 16 }}
              rowKey={(record) => record.id}
              dataSource={sessionMessages}
              pagination={false}
              columns={[
                { title: 'Role', dataIndex: 'role' },
                { title: 'Content', dataIndex: 'content' },
                { title: 'Confidence', dataIndex: 'confidence' },
                { title: 'Created At', dataIndex: 'created_at' },
              ]}
            />
          </>
        ) : (
          <Alert type="info" message="暂无会话信息，请先获取会话。" showIcon />
        )}
      </Card>
    </div>
  )
}
