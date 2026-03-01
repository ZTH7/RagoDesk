import { Button, Card, Input, Space, Typography } from 'antd'
import { useEffect, useMemo, useState } from 'react'
import { Link, useNavigate, useParams } from 'react-router-dom'
import { publicChatApi } from '../../services/publicChat'
import { uiMessage } from '../../services/uiMessage'

type ChatMessage = {
  id: string
  role: 'user' | 'assistant'
  content: string
}

function getVisitorID() {
  const key = 'ragodesk.public.visitor_id'
  const existing = window.localStorage.getItem(key)
  if (existing) return existing
  const next = `visitor_${Math.random().toString(36).slice(2, 10)}`
  window.localStorage.setItem(key, next)
  return next
}

export function PublicChat() {
  const { chatId = '', sessionId: routeSessionID = '' } = useParams()
  const navigate = useNavigate()
  const resolvedChatID = (() => {
    try {
      return decodeURIComponent(chatId)
    } catch {
      return chatId
    }
  })()
  const [messages, setMessages] = useState<ChatMessage[]>([])
  const [loading, setLoading] = useState(false)
  const [input, setInput] = useState('')
  const [sessionID, setSessionID] = useState(routeSessionID)
  const [initialized, setInitialized] = useState(false)

  useEffect(() => {
    setSessionID(routeSessionID || '')
    setMessages([])
    setInitialized(false)
  }, [routeSessionID, resolvedChatID])

  useEffect(() => {
    if (!resolvedChatID || !routeSessionID || initialized) return
    let active = true
    setLoading(true)
    publicChatApi
      .getSession(resolvedChatID, routeSessionID)
      .then((res) => {
        if (!active) return
        setMessages(
          (res.messages || [])
            .filter((m) => m.role === 'user' || m.role === 'assistant')
            .map((m) => ({
              id: m.id || `${m.role}_${Math.random().toString(36).slice(2, 8)}`,
              role: m.role as 'user' | 'assistant',
              content: m.content,
            })),
        )
        setInitialized(true)
      })
      .catch((err: Error) => {
        if (!active) return
        uiMessage.error(err.message)
      })
      .finally(() => {
        if (!active) return
        setLoading(false)
      })
    return () => {
      active = false
    }
  }, [resolvedChatID, routeSessionID, initialized])

  const canSend = useMemo(() => !!input.trim() && !!resolvedChatID && !loading, [input, resolvedChatID, loading])

  const ensureSession = async () => {
    if (sessionID) return sessionID
    const res = await publicChatApi.createSession(resolvedChatID, getVisitorID())
    const created = res.session?.id
    if (!created) throw new Error('创建会话失败')
    setSessionID(created)
    navigate(`/chat/${encodeURIComponent(resolvedChatID)}/${created}`, { replace: true })
    return created
  }

  const handleSend = async () => {
    const text = input.trim()
    if (!text) return
    if (!resolvedChatID) {
      uiMessage.error('链接无效：缺少聊天标识')
      return
    }
    try {
      setLoading(true)
      setInput('')
      const localUserMsg: ChatMessage = {
        id: `local_user_${Date.now()}`,
        role: 'user',
        content: text,
      }
      setMessages((prev) => [...prev, localUserMsg])
      const sid = await ensureSession()
      const res = await publicChatApi.sendMessage(resolvedChatID, sid, text)
      const reply: ChatMessage = {
        id: `local_assistant_${Date.now()}`,
        role: 'assistant',
        content: res.reply || '抱歉，暂时无法回答，请稍后再试。',
      }
      setMessages((prev) => [...prev, reply])
    } catch (err) {
      const msg = err instanceof Error ? err.message : '发送失败'
      uiMessage.error(msg)
    } finally {
      setLoading(false)
    }
  }

  return (
    <div style={{ minHeight: '100vh', padding: 24, maxWidth: 960, margin: '0 auto' }}>
      <Card>
        <Space direction="vertical" size="middle" style={{ width: '100%' }}>
          <Space style={{ justifyContent: 'space-between', width: '100%' }}>
            <Space direction="vertical" size={0}>
              <Typography.Title level={4} style={{ margin: 0 }}>
                在线客服
              </Typography.Title>
              <Typography.Text className="muted">
                输入问题后将自动创建会话，并应用当前链接对应的 API Key 限流与统计规则。
              </Typography.Text>
            </Space>
            <Link to="/">返回首页</Link>
          </Space>

          {sessionID ? (
            <Typography.Text className="muted">当前会话：{sessionID}</Typography.Text>
          ) : (
            <Typography.Text className="muted">当前会话：尚未开始（发送第一条消息后创建）</Typography.Text>
          )}

          <div
            style={{
              border: '1px solid var(--ant-color-border)',
              borderRadius: 12,
              padding: 16,
              minHeight: 360,
              maxHeight: 520,
              overflowY: 'auto',
              background: 'var(--ant-color-bg-layout)',
            }}
          >
            {messages.length === 0 ? (
              <Typography.Text className="muted">您好，请输入问题开始对话。</Typography.Text>
            ) : null}
            <Space direction="vertical" size="middle" style={{ width: '100%' }}>
              {messages.map((m) => (
                <div
                  key={m.id}
                  style={{
                    alignSelf: m.role === 'user' ? 'flex-end' : 'flex-start',
                    background:
                      m.role === 'user'
                        ? 'var(--ant-color-primary-bg)'
                        : 'var(--ant-color-bg-container)',
                    border: '1px solid var(--ant-color-border-secondary)',
                    borderRadius: 10,
                    padding: '8px 12px',
                    maxWidth: '85%',
                  }}
                >
                  <Typography.Text>{m.content}</Typography.Text>
                </div>
              ))}
            </Space>
          </div>

          <Input.TextArea
            value={input}
            onChange={(e) => setInput(e.target.value)}
            rows={4}
            placeholder="请输入您的问题..."
            onPressEnter={(e) => {
              if (!e.shiftKey) {
                e.preventDefault()
                handleSend()
              }
            }}
          />
          <Space style={{ justifyContent: 'space-between', width: '100%' }}>
            <Typography.Text className="muted">回车发送，Shift + 回车换行</Typography.Text>
            <Button type="primary" loading={loading} disabled={!canSend} onClick={handleSend}>
              发送
            </Button>
          </Space>
        </Space>
      </Card>
    </div>
  )
}
