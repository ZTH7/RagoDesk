import { Layout, Menu, Avatar, Dropdown, Space, Typography, Tag, Button } from 'antd'
import { useEffect, useMemo, useState } from 'react'
import { LogoutOutlined, UserOutlined } from '@ant-design/icons'
import { Outlet, useLocation, useNavigate } from 'react-router-dom'
import { consoleNavItems, consoleMenuKeys } from '../routes/console'
import { buildMenuItems, resolveSelectedKey } from '../routes/utils'
import { usePermissions } from '../auth/PermissionContext'
import { getTenantId, getToken } from '../auth/storage'

const { Sider, Header, Content } = Layout

export function ConsoleLayout() {
  const location = useLocation()
  const navigate = useNavigate()
  const { permissions } = usePermissions()
  const selectedKey = resolveSelectedKey(location.pathname, consoleMenuKeys)
  const [openKeys, setOpenKeys] = useState<string[]>([])
  const tenantId = getTenantId()
  const token = getToken()

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
      <Sider width={240} theme="light">
        <div className="app-logo">RagoDesk</div>
        <Menu
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
            <Typography.Text className="muted">Tenant: {tenantId || '-'}</Typography.Text>
          </Space>
          <Space align="center">
            <Button size="small" type="default">
              中文
            </Button>
            <Dropdown
              menu={{
                items: [
                  { key: 'profile', label: '个人中心' },
                  { key: 'logout', label: '退出登录', icon: <LogoutOutlined /> },
                ],
                onClick: ({ key }) => {
                  if (key === 'profile') {
                    navigate('/console/profile')
                  }
                },
              }}
            >
              <Space style={{ cursor: 'pointer' }}>
                <Avatar icon={<UserOutlined />} />
                <Typography.Text>{token ? '已登录' : '未登录'}</Typography.Text>
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
