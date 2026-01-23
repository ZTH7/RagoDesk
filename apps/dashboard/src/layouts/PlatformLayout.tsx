import { Layout, Menu, Avatar, Dropdown, Space, Typography, Tag } from 'antd'
import { LogoutOutlined, UserOutlined } from '@ant-design/icons'
import { Outlet, useLocation, useNavigate } from 'react-router-dom'
import { platformNavItems, platformMenuKeys } from '../routes/platform'
import { buildMenuItems, resolveSelectedKey } from '../routes/utils'
import { usePermissions } from '../auth/PermissionContext'
import { useMemo } from 'react'
import { getToken } from '../auth/storage'

const { Sider, Header, Content } = Layout

export function PlatformLayout() {
  const location = useLocation()
  const navigate = useNavigate()
  const { permissions } = usePermissions()
  const selectedKey = resolveSelectedKey(location.pathname, platformMenuKeys)
  const menuItems = useMemo(() => buildMenuItems(platformNavItems, permissions), [permissions])
  const token = getToken()

  return (
    <Layout className="app-shell">
      <Sider width={240} theme="light">
        <div className="app-logo">RagoDesk</div>
        <Menu
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
            <Typography.Text className="muted">Platform Scope</Typography.Text>
          </Space>
          <Dropdown
            menu={{
              items: [
                { key: 'profile', label: '个人中心' },
                { key: 'logout', label: '退出登录', icon: <LogoutOutlined /> },
              ],
            }}
          >
            <Space style={{ cursor: 'pointer' }}>
              <Avatar icon={<UserOutlined />} />
              <Typography.Text>{token ? '已登录' : '未登录'}</Typography.Text>
            </Space>
          </Dropdown>
        </Header>
        <Content className="app-content">
          <Outlet />
        </Content>
      </Layout>
    </Layout>
  )
}
