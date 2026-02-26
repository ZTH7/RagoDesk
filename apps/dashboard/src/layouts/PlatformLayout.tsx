import { Layout, Menu, Avatar, Dropdown, Space, Typography, Tag } from 'antd'
import { LogoutOutlined, UserOutlined } from '@ant-design/icons'
import { Outlet, useLocation, useNavigate } from 'react-router-dom'
import { platformNavItems, platformMenuKeys } from '../routes/platform'
import { buildMenuItems, resolveSelectedKey } from '../routes/utils'
import { usePermissions } from '../auth/PermissionContext'
import { useMemo } from 'react'
import { clearProfile, clearScope, clearToken, getProfile, getToken } from '../auth/storage'
import { ThemeModeToggle, ThemeStatusDot } from '../components/ThemeModeToggle'
import { useThemeMode } from '../theme/mode'

const { Sider, Header, Content } = Layout

export function PlatformLayout() {
  const location = useLocation()
  const navigate = useNavigate()
  const { permissions } = usePermissions()
  const selectedKey = resolveSelectedKey(location.pathname, platformMenuKeys)
  const menuItems = useMemo(() => buildMenuItems(platformNavItems, permissions), [permissions])
  const token = getToken()
  const profile = getProfile()
  const displayName = profile?.name || profile?.account || (token ? '已登录' : '未登录')
  const { resolvedMode } = useThemeMode()
  const siderTheme = resolvedMode === 'dark' ? 'dark' : 'light'

  return (
    <Layout className="app-shell">
      <Sider width={240} theme={siderTheme}>
        <div className="app-logo">RagoDesk</div>
        <Menu
          theme={siderTheme}
          mode="inline"
          items={menuItems}
          selectedKeys={[selectedKey]}
          onClick={(info) => {
            if (typeof info.key === 'string' && info.key.startsWith('/')) {
              navigate(info.key)
            }
          }}
        />
      </Sider>
      <Layout>
        <Header className="app-header">
          <Space align="center">
            <Tag color="purple">Platform</Tag>
            <ThemeStatusDot />
            <Typography.Text className="muted">Platform Scope</Typography.Text>
          </Space>
          <Space align="center">
            <ThemeModeToggle />
            <Dropdown
              menu={{
                items: [
                  { key: 'profile', label: '个人中心' },
                  { key: 'logout', label: '退出登录', icon: <LogoutOutlined /> },
                ],
                onClick: ({ key }) => {
                  if (key === 'logout') {
                    clearToken()
                    clearProfile()
                    clearScope()
                    navigate('/')
                  }
                },
              }}
            >
              <Space style={{ cursor: 'pointer' }}>
                <Avatar icon={<UserOutlined />} />
                <Typography.Text>{displayName}</Typography.Text>
              </Space>
            </Dropdown>
          </Space>
        </Header>
        <Content className="app-content">
          <Outlet />
        </Content>
      </Layout>
    </Layout>
  )
}
