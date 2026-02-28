import { Button, Card, Descriptions, Space, Tag, Typography } from 'antd'
import { useMemo } from 'react'
import { useNavigate } from 'react-router-dom'
import { PageHeader } from '../../components/PageHeader'
import {
  clearProfile,
  clearScope,
  clearTenantId,
  clearToken,
  getCurrentTenantId,
  getProfile,
  getToken,
} from '../../auth/storage'

import { uiMessage } from '../../services/uiMessage'
export function Profile() {
  const navigate = useNavigate()
  const profile = getProfile()
  const token = getToken()
  const tenantId = getCurrentTenantId()
  const roleText = useMemo(() => (profile?.roles?.length ? profile.roles.join(', ') : '-'), [profile?.roles])

  return (
    <div className="page">
      <PageHeader title="个人中心" description="当前账号与会话信息" />
      <Card>
        <Descriptions column={1} bordered size="middle">
          <Descriptions.Item label="账号">{profile?.account || '-'}</Descriptions.Item>
          <Descriptions.Item label="姓名">{profile?.name || '-'}</Descriptions.Item>
          <Descriptions.Item label="Scope">
            <Tag color={profile?.scope === 'platform' ? 'purple' : 'blue'}>
              {profile?.scope || '-'}
            </Tag>
          </Descriptions.Item>
          <Descriptions.Item label="Tenant">{tenantId || '-'}</Descriptions.Item>
          <Descriptions.Item label="角色">{roleText}</Descriptions.Item>
          <Descriptions.Item label="Token 状态">{token ? '已登录' : '未登录'}</Descriptions.Item>
        </Descriptions>
      </Card>
      <Card>
        <Typography.Paragraph type="secondary" style={{ marginBottom: 16 }}>
          账户信息由登录态自动管理。若需切换账号，请先退出再重新登录。
        </Typography.Paragraph>
        <Space>
          <Button
            onClick={() => {
              clearToken()
              clearTenantId()
              clearProfile()
              clearScope()
              uiMessage.success('已退出登录')
              navigate('/')
            }}
          >
            退出登录
          </Button>
        </Space>
      </Card>
    </div>
  )
}

