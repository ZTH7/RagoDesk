import { Avatar, Button, Card, Col, Row, Space, Tag, Typography } from 'antd'
import { useMemo } from 'react'
import { useNavigate } from 'react-router-dom'
import {
  clearProfile,
  clearScope,
  clearTenantId,
  clearToken,
  getProfile,
  getScope,
  getToken,
} from '../auth/storage'

export function Home() {
  const navigate = useNavigate()
  const token = getToken()
  const profile = getProfile()
  const scope = getScope()

  const displayName = useMemo(() => {
    return profile?.name || profile?.account || (token ? '已登录用户' : '未登录')
  }, [profile?.name, profile?.account, token])

  const avatarText = useMemo(() => {
    if (!displayName) return 'R'
    return displayName.slice(0, 1).toUpperCase()
  }, [displayName])

  const handleLogout = () => {
    clearToken()
    clearTenantId()
    clearProfile()
    clearScope()
    navigate('/')
  }

  return (
    <div className="home-shell">
      <section className="home-hero">
        <Typography.Title level={2}>RagoDesk</Typography.Title>
        <Typography.Paragraph className="muted">
          多租户 AI 客服平台，覆盖知识库、RAG、会话与 API 管理的全链路。
        </Typography.Paragraph>
        <Space size="middle">
          <Button type="primary" onClick={() => navigate('/console/login')}>
            Console 登录
          </Button>
          <Button onClick={() => navigate('/console/register')}>Console 注册</Button>
          <Button type="dashed" onClick={() => navigate('/platform/login')}>
            Platform 登录
          </Button>
        </Space>
      </section>

      <Row gutter={16} className="home-cards">
        {token ? (
          <Col xs={24} lg={12}>
            <Card>
              <Space direction="vertical" size="middle" style={{ width: '100%' }}>
                <Space align="center">
                  <Avatar size={56}>{avatarText}</Avatar>
                  <div>
                    <Typography.Title level={4} style={{ margin: 0 }}>
                      {displayName}
                    </Typography.Title>
                    <Space>
                      <Tag color="blue">{scope || 'console'}</Tag>
                      {profile?.roles?.map((role) => (
                        <Tag key={role}>{role}</Tag>
                      ))}
                    </Space>
                  </div>
                </Space>
                <Space wrap>
                  {scope !== 'platform' && (
                    <Button type="primary" onClick={() => navigate('/console/analytics/overview')}>
                      进入 Console
                    </Button>
                  )}
                  {scope !== 'console' && (
                    <Button type="primary" onClick={() => navigate('/platform/tenants')}>
                      进入 Platform
                    </Button>
                  )}
                  <Button onClick={handleLogout}>退出登录</Button>
                </Space>
              </Space>
            </Card>
          </Col>
        ) : null}

        <Col xs={24} lg={12}>
          <Card title="标准使用流程">
            <ol className="home-list">
              <li>注册租户并创建管理员</li>
              <li>创建知识库并上传文档</li>
              <li>创建 Bot 并绑定知识库</li>
              <li>创建 API Key 并开始调用</li>
            </ol>
          </Card>
        </Col>
      </Row>

      <Row gutter={16}>
        <Col xs={24} md={8}>
          <Card title="Console 区">
            <Typography.Paragraph className="muted">
              租户侧的日常运营入口，包含知识库、机器人、会话、调用统计等能力。
            </Typography.Paragraph>
            <Button type="primary" onClick={() => navigate('/console/login')}>
              进入 Console
            </Button>
          </Card>
        </Col>
        <Col xs={24} md={8}>
          <Card title="Platform 区">
            <Typography.Paragraph className="muted">
              平台运维入口，管理租户、平台管理员、权限与角色。
            </Typography.Paragraph>
            <Button onClick={() => navigate('/platform/login')}>进入 Platform</Button>
          </Card>
        </Col>
        <Col xs={24} md={8}>
          <Card title="API 调试">
            <Typography.Paragraph className="muted">
              Console 登录后可通过 API 调试页验证会话与 RAG 调用。
            </Typography.Paragraph>
            <Button onClick={() => navigate('/console/devtools/api')}>打开调试</Button>
          </Card>
        </Col>
      </Row>
    </div>
  )
}
