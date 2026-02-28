import { Avatar, Button, Card, Col, Row, Space, Statistic, Tag, Typography } from 'antd'
import { useMemo } from 'react'
import { useNavigate } from 'react-router-dom'
import { ThemeModeToggle, ThemeStatusDot } from '../components/ThemeModeToggle'
import {
  BarChartOutlined,
  CustomerServiceOutlined,
  GlobalOutlined,
  RocketOutlined,
  SafetyCertificateOutlined,
  ThunderboltOutlined,
} from '@ant-design/icons'
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

  const featureCards = [
    {
      title: '快速上线',
      desc: '从创建账号到对外提供机器人服务，按标准流程即可快速完成。',
      icon: <RocketOutlined />,
    },
    {
      title: '更稳定回答',
      desc: '围绕企业知识构建回答链路，减少答非所问与信息漂移。',
      icon: <CustomerServiceOutlined />,
    },
    {
      title: '多渠道接入',
      desc: '通过统一接口接入官网、APP、小程序等多种业务入口。',
      icon: <GlobalOutlined />,
    },
    {
      title: '可运营可分析',
      desc: '沉淀调用与会话数据，帮助团队持续优化机器人效果。',
      icon: <BarChartOutlined />,
    },
    {
      title: '安全可控',
      desc: '通过权限体系和操作边界，保障组织内外协作的安全性。',
      icon: <SafetyCertificateOutlined />,
    },
    {
      title: '服务体验提升',
      desc: '让用户问题更快得到响应，提升客服效率与用户满意度。',
      icon: <ThunderboltOutlined />,
    },
  ]

  const handleLogout = () => {
    clearToken()
    clearTenantId()
    clearProfile()
    clearScope()
    navigate('/')
  }

  return (
    <div className="home-shell motion-enter">
      <section className="home-nav">
        <div className="home-brand">
          <Space>
            <Typography.Text strong>RagoDesk</Typography.Text>
            <ThemeStatusDot />
            <Tag color="blue">AI Customer Support</Tag>
          </Space>
        </div>
        <div className="home-nav-actions">
          <ThemeModeToggle />
        </div>
      </section>

      <section className="home-hero">
        <div className="home-hero-copy">
          <Tag color="cyan" className="home-badge">
            Enterprise AI Support Platform
          </Tag>
          <Typography.Title level={1} className="home-title">
            让企业知识，稳定驱动每一次 AI 客服对话
          </Typography.Title>
          <Typography.Paragraph className="muted home-subtitle">
            RagoDesk 是面向业务团队的 AI 客服工作台：你可以创建机器人、管理知识内容、
            接入业务渠道，并通过数据看板持续优化回答质量与服务体验。
          </Typography.Paragraph>
          <Space size="middle" wrap>
            {!token && (
              <>
                <Button type="primary" size="large" onClick={() => navigate('/console/register')}>
                  立即开始
                </Button>
                <Button size="large" onClick={() => navigate('/console/login')}>
                  Console 登录
                </Button>
                <Button type="dashed" size="large" onClick={() => navigate('/platform/login')}>
                  Platform 登录
                </Button>
              </>
            )}
            {token && (
              <>
                {scope !== 'platform' && (
                  <Button type="primary" size="large" onClick={() => navigate('/console/analytics/overview')}>
                    进入 Console
                  </Button>
                )}
                {scope !== 'console' && (
                  <Button type="primary" size="large" onClick={() => navigate('/platform/tenants')}>
                    进入 Platform
                  </Button>
                )}
                <Button size="large" onClick={handleLogout}>
                  退出登录
                </Button>
              </>
            )}
          </Space>
        </div>
        <Card className="home-hero-panel surface-card" bordered={false}>
          <Space direction="vertical" size="middle" style={{ width: '100%' }}>
            <Space align="center">
              <Avatar size={56}>{avatarText}</Avatar>
              <div>
                <Typography.Title level={4} style={{ margin: 0 }}>
                  {displayName}
                </Typography.Title>
                <Space wrap>
                  <Tag color={scope === 'platform' ? 'purple' : 'blue'}>{scope || 'guest'}</Tag>
                  {profile?.roles?.slice(0, 2).map((role) => (
                    <Tag key={role}>{role}</Tag>
                  ))}
                </Space>
              </div>
            </Space>
            <Row gutter={12}>
              <Col span={12}>
                <Statistic title="入门流程" value="4 步完成" />
              </Col>
              <Col span={12}>
                <Statistic title="主要场景" value="客服 / 咨询 / 支持" />
              </Col>
            </Row>
            <Typography.Text className="muted">
              推荐先完成：创建知识库 → 上传文档 → 配置机器人 → 对外发布
            </Typography.Text>
          </Space>
        </Card>
      </section>

      <section className="home-metrics-grid">
        <Card className="surface-card">
          <Statistic title="统一知识入口" value="文档集中管理" />
        </Card>
        <Card className="surface-card">
          <Statistic title="机器人配置" value="按业务灵活组合" />
        </Card>
        <Card className="surface-card">
          <Statistic title="会话闭环" value="问题与回答可追踪" />
        </Card>
        <Card className="surface-card">
          <Statistic title="运营分析" value="趋势与缺口一屏可见" />
        </Card>
      </section>

      <section className="home-feature-grid">
        {featureCards.map((item, idx) => (
          <Card
            key={item.title}
            className="surface-card home-feature-card"
            bordered={false}
            style={{ animationDelay: `${80 * idx}ms` }}
          >
            <Space direction="vertical" size="small">
              <span className="home-feature-icon">{item.icon}</span>
              <Typography.Title level={5} style={{ margin: 0 }}>
                {item.title}
              </Typography.Title>
              <Typography.Paragraph className="muted" style={{ marginBottom: 0 }}>
                {item.desc}
              </Typography.Paragraph>
            </Space>
          </Card>
        ))}
      </section>

      <Row gutter={16}>
        <Col xs={24} md={8}>
          <Card className="surface-card" bordered={false} title="Console 区">
            <Typography.Paragraph className="muted">
              租户侧的日常运营入口，包含知识库、机器人、会话、调用统计等能力。
            </Typography.Paragraph>
            <Button type="primary" onClick={() => navigate('/console/login')}>
              进入 Console
            </Button>
          </Card>
        </Col>
        <Col xs={24} md={8}>
          <Card className="surface-card" bordered={false} title="Platform 区">
            <Typography.Paragraph className="muted">
              平台运维入口，管理租户、平台管理员、权限与角色。
            </Typography.Paragraph>
            <Button onClick={() => navigate('/platform/login')}>进入 Platform</Button>
          </Card>
        </Col>
        <Col xs={24} md={8}>
          <Card className="surface-card" bordered={false} title="开发与接入">
            <Typography.Paragraph className="muted">
              Console 登录后可通过调试页面验证会话调用，再接入业务系统。
            </Typography.Paragraph>
            <Button onClick={() => navigate('/console/devtools/api')}>进入调试</Button>
          </Card>
        </Col>
      </Row>

      <section className="home-flow">
        <Typography.Title level={4} style={{ marginBottom: 12 }}>
          标准落地流程
        </Typography.Title>
        <Row gutter={[12, 12]}>
          {[
            '创建知识库并定义业务范围',
            '上传文档，自动解析与向量化',
            '创建 Bot 并绑定知识库',
            '创建 API Key 并接入业务系统',
          ].map((step, idx) => (
            <Col xs={24} md={12} lg={6} key={step}>
              <Card className="surface-card home-step-card" bordered={false}>
                <Typography.Text strong>STEP {idx + 1}</Typography.Text>
                <Typography.Paragraph style={{ marginBottom: 0 }}>{step}</Typography.Paragraph>
              </Card>
            </Col>
          ))}
        </Row>
      </section>
    </div>
  )
}
