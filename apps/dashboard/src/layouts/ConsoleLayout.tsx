import { Button, Layout, Menu, Avatar, Dropdown, Space, Typography, Tag } from 'antd'
import { useEffect, useMemo, useState } from 'react'
import { LogoutOutlined, UserOutlined } from '@ant-design/icons'
import { Link, Outlet, useLocation, useNavigate } from 'react-router-dom'
import { consoleNavItems, consoleMenuKeys } from '../routes/console'
import { buildMenuItems, resolveSelectedKey } from '../routes/utils'
import { usePermissions } from '../auth/PermissionContext'
import {
  clearProfile,
  clearScope,
  clearTenantId,
  clearToken,
  getCurrentTenantId,
  getProfile,
  getToken,
} from '../auth/storage'
import { ThemeModeToggle, ThemeStatusDot } from '../components/ThemeModeToggle'
import { useThemeMode } from '../theme/mode'
import { RequestBanner } from '../components/RequestBanner'

const { Sider, Header, Content } = Layout

export function ConsoleLayout() {
  const location = useLocation()
  const navigate = useNavigate()
  const { permissions, error, stale, refresh, loading: permissionLoading } = usePermissions()
  const selectedKey = resolveSelectedKey(location.pathname, consoleMenuKeys)
  const [openKeys, setOpenKeys] = useState<string[]>([])
  const tenantId = getCurrentTenantId()
  const token = getToken()
  const profile = getProfile()
  const displayName = profile?.name || profile?.account || (token ? '已登录' : '未登录')
  const { resolvedMode } = useThemeMode()
  const siderTheme = resolvedMode === 'dark' ? 'dark' : 'light'

  useEffect(() => {
    if (selectedKey.startsWith('/console/analytics')) {
      setOpenKeys(['/console/analytics'])
      return
    }
    setOpenKeys([])
  }, [selectedKey])

  const menuItems = useMemo(() => buildMenuItems(consoleNavItems, permissions), [permissions])

  return (
    <Layout className="app-shell">
      <Sider width={240} theme={siderTheme}>
        <div className="app-logo">
          <Link to="/" className="app-logo-link">
            RagoDesk
          </Link>
        </div>
        <Menu
          theme={siderTheme}
          mode="inline"
          items={menuItems}
          selectedKeys={[selectedKey]}
          openKeys={openKeys}
          onOpenChange={(keys) => setOpenKeys(keys as string[])}
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
            <Tag color="blue">Console</Tag>
            <ThemeStatusDot />
            <Typography.Text className="muted">Tenant: {tenantId || '-'}</Typography.Text>
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
                  if (key === 'profile') {
                    navigate('/console/profile')
                    return
                  }
                  if (key === 'logout') {
                    clearToken()
                    clearTenantId()
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
          <RequestBanner
            error={error}
            title={stale ? '权限同步失败，当前使用本地缓存权限' : '权限加载失败'}
            action={
              <Button size="small" onClick={refresh} loading={permissionLoading}>
                重试
              </Button>
            }
          />
          <Outlet />
        </Content>
      </Layout>
    </Layout>
  )
}
