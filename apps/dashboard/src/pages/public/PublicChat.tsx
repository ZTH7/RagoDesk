import { RobotOutlined, SendOutlined, UserOutlined } from '@ant-design/icons'
import { Avatar, Button, Card, Input, Space, Typography } from 'antd'
import { useEffect, useMemo, useRef, useState } from 'react'
import { Link, useNavigate, useParams } from 'react-router-dom'
import { publicChatApi } from '../../services/publicChat'
import { uiMessage } from '../../services/uiMessage'
import './PublicChat.css'

type ChatMessage = {
  id: string
  role: 'user' | 'assistant'
  content: string
  timestamp?: string
}

function getVisitorID() {
  const key = 'ragodesk.public.visitor_id'
  const existing = window.localStorage.getItem(key)
  if (existing) return existing
  const next = `visitor_${Math.random().toString(36).slice(2, 10)}`
  window.localStorage.setItem(key, next)
  return next
}

function formatNow() {
  const now = new Date()
  return `${String(now.getHours()).padStart(2, '0')}:${String(now.getMinutes()).padStart(2, '0')}`
}

function formatCreatedAt(ts?: { seconds?: number; nanos?: number }) {
  if (!ts?.seconds) return undefined
  const ms = ts.seconds * 1000 + Math.floor((ts.nanos ?? 0) / 1_000_000)
  const dt = new Date(ms)
  return `${String(dt.getHours()).padStart(2, '0')}:${String(dt.getMinutes()).padStart(2, '0')}`
}

export function PublicChat() {
  const { chatId = '', sessionId: routeSessionID = '' } = useParams()
  const navigate = useNavigate()
  const resolvedChatID = useMemo(() => {
    try {
      return decodeURIComponent(chatId)
    } catch {
      return chatId
    }
  }, [chatId])

  const [messages, setMessages] = useState<ChatMessage[]>([])
  const [input, setInput] = useState('')
  const [sessionID, setSessionID] = useState(routeSessionID)
  const [hydrating, setHydrating] = useState(false)
  const [sending, setSending] = useState(false)

  const prevChatIDRef = useRef(resolvedChatID)
  const skipHydrateForSessionRef = useRef('')
  const hydratedSessionIDRef = useRef('')
  const messageContainerRef = useRef<HTMLDivElement | null>(null)

  useEffect(() => {
    if (prevChatIDRef.current === resolvedChatID) return
    prevChatIDRef.current = resolvedChatID
    skipHydrateForSessionRef.current = ''
    hydratedSessionIDRef.current = ''
    setSessionID(routeSessionID || '')
    setMessages([])
    setHydrating(false)
    setSending(false)
    setInput('')
  }, [resolvedChatID, routeSessionID])

  useEffect(() => {
    if (!routeSessionID) {
      if (!sessionID) return
      setSessionID('')
      hydratedSessionIDRef.current = ''
      setMessages([])
      return
    }
    if (routeSessionID === sessionID) return
    setSessionID(routeSessionID)
    if (skipHydrateForSessionRef.current === routeSessionID) {
      skipHydrateForSessionRef.current = ''
      hydratedSessionIDRef.current = routeSessionID
      return
    }
    hydratedSessionIDRef.current = ''
    setMessages([])
  }, [routeSessionID, sessionID])

  useEffect(() => {
    if (!resolvedChatID || !sessionID) return
    if (hydratedSessionIDRef.current === sessionID) return
    let active = true
    setHydrating(true)
    publicChatApi
      .getSession(resolvedChatID, sessionID)
      .then((res) => {
        if (!active) return
        setMessages(
          (res.messages || [])
            .filter((m) => m.role === 'user' || m.role === 'assistant')
            .map((m) => ({
              id: m.id || `${m.role}_${Math.random().toString(36).slice(2, 8)}`,
              role: m.role as 'user' | 'assistant',
              content: m.content,
              timestamp: formatCreatedAt(m.created_at),
            })),
        )
        hydratedSessionIDRef.current = sessionID
      })
      .catch((err: Error) => {
        if (!active) return
        uiMessage.error(err.message)
      })
      .finally(() => {
        if (!active) return
        setHydrating(false)
      })
    return () => {
      active = false
    }
  }, [resolvedChatID, sessionID])

  useEffect(() => {
    const el = messageContainerRef.current
    if (!el) return
    el.scrollTo({ top: el.scrollHeight, behavior: 'smooth' })
  }, [messages, sending])

  const canSend = useMemo(
    () => !!input.trim() && !!resolvedChatID && !sending && !hydrating,
    [hydrating, input, resolvedChatID, sending],
  )

  const ensureSession = async () => {
    if (sessionID) return sessionID
    const res = await publicChatApi.createSession(resolvedChatID, getVisitorID())
    const created = res.session?.id
    if (!created) throw new Error('创建会话失败')
    skipHydrateForSessionRef.current = created
    hydratedSessionIDRef.current = created
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
    const localUserMsg: ChatMessage = {
      id: `local_user_${Date.now()}`,
      role: 'user',
      content: text,
      timestamp: formatNow(),
    }
    setMessages((prev) => [...prev, localUserMsg])
    setInput('')
    try {
      setSending(true)
      const sid = await ensureSession()
      const res = await publicChatApi.sendMessage(resolvedChatID, sid, text)
      const reply: ChatMessage = {
        id: `local_assistant_${Date.now()}`,
        role: 'assistant',
        content: res.reply || '抱歉，暂时无法回答，请稍后再试。',
        timestamp: formatNow(),
      }
      setMessages((prev) => [...prev, reply])
    } catch (err) {
      const msg = err instanceof Error ? err.message : '发送失败'
      setMessages((prev) => [
        ...prev,
        {
          id: `local_error_${Date.now()}`,
          role: 'assistant',
          content: `抱歉，当前服务繁忙：${msg}`,
          timestamp: formatNow(),
        },
      ])
      uiMessage.error(msg)
    } finally {
      setSending(false)
    }
  }

  return (
    <div className="public-chat-page">
      <Card className="public-chat-card" bordered={false}>
        <div className="public-chat-header">
          <div>
            <Typography.Title level={4} style={{ margin: 0 }}>
              在线客服
            </Typography.Title>
            <Typography.Text className="muted">
              你好，我是你的智能助手。请输入问题，我会尽快给你答案。
            </Typography.Text>
          </div>
          <Link to="/">返回首页</Link>
        </div>

        <Typography.Text className="muted" style={{ marginBottom: 12 }}>
          {sessionID ? `会话编号：${sessionID}` : '会话状态：发送第一条消息后自动创建'}
        </Typography.Text>

        <div ref={messageContainerRef} className="public-chat-message-panel">
          {messages.length === 0 && !hydrating ? (
            <div className="public-chat-empty">
              <RobotOutlined style={{ fontSize: 24 }} />
              <Typography.Text>欢迎使用在线客服，请直接输入你的问题。</Typography.Text>
            </div>
          ) : null}

          {messages.map((m) => (
            <div
              key={m.id}
              className={`public-chat-row ${m.role === 'user' ? 'is-user' : 'is-assistant'} public-chat-enter`}
            >
              <Avatar
                size={32}
                className={`public-chat-avatar ${m.role === 'user' ? 'is-user' : 'is-assistant'}`}
                icon={m.role === 'user' ? <UserOutlined /> : <RobotOutlined />}
              />
              <div className="public-chat-bubble-wrap">
                <div className={`public-chat-bubble ${m.role === 'user' ? 'is-user' : 'is-assistant'}`}>
                  {m.content}
                </div>
                {m.timestamp ? <span className="public-chat-time">{m.timestamp}</span> : null}
              </div>
            </div>
          ))}

          {sending ? (
            <div className="public-chat-row is-assistant public-chat-enter">
              <Avatar size={32} className="public-chat-avatar is-assistant" icon={<RobotOutlined />} />
              <div className="public-chat-bubble-wrap">
                <div className="public-chat-bubble is-assistant">
                  <span className="public-chat-typing-dot" />
                  <span className="public-chat-typing-dot" />
                  <span className="public-chat-typing-dot" />
                </div>
              </div>
            </div>
          ) : null}
        </div>

        <div className="public-chat-input-box">
          <Input.TextArea
            value={input}
            onChange={(e) => setInput(e.target.value)}
            rows={3}
            placeholder="请输入你的问题，按 Enter 发送（Shift + Enter 换行）"
            onPressEnter={(e) => {
              if (!e.shiftKey) {
                e.preventDefault()
                handleSend()
              }
            }}
          />
          <Space style={{ justifyContent: 'space-between', width: '100%', marginTop: 10 }}>
            <Typography.Text className="muted">
              仅用于客服问答，不建议输入敏感信息。
            </Typography.Text>
            <Button type="primary" icon={<SendOutlined />} loading={sending} disabled={!canSend} onClick={handleSend}>
              发送
            </Button>
          </Space>
        </div>
      </Card>
    </div>
  )
}
