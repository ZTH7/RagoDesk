import { Button, Card, Descriptions, Empty, Modal, Popconfirm, Select, Space, Tag, Skeleton, Typography } from 'antd'
import { useState } from 'react'
import { useParams } from 'react-router-dom'
import { PageHeader } from '../../components/PageHeader'
import { TechnicalMeta } from '../../components/TechnicalMeta'
import { RequestBanner } from '../../components/RequestBanner'
import { useRequest } from '../../hooks/useRequest'
import { platformApi } from '../../services/platform'
import { formatDateTime } from '../../utils/datetime'

import { uiMessage } from '../../services/uiMessage'
export function PlatformAdminDetail() {
  const { id } = useParams()
  const adminId = id ?? ''
  const [assignOpen, setAssignOpen] = useState(false)
  const [selectedRole, setSelectedRole] = useState('')

  const adminRequest = useRequest(
    () => platformApi.getAdmin(adminId),
    { admin: { id: '', email: '', name: '', status: '', created_at: '' } },
    { enabled: Boolean(adminId), deps: [adminId] },
  )
  const { data, loading } = adminRequest
  const admin = data.admin
  const { data: roleData } = useRequest(() => platformApi.listRoles(), { items: [] })
  const adminRolesRequest = useRequest(
    () => platformApi.listAdminRoles(adminId),
    { items: [] },
    { enabled: Boolean(adminId), deps: [adminId] },
  )
  const assignedRoles = adminRolesRequest.data.items
  const requestError = adminRequest.error || adminRolesRequest.error

  const handleAssign = async () => {
    if (!selectedRole) {
      uiMessage.error('请选择角色')
      return
    }
    try {
      await platformApi.assignAdminRole(adminId, selectedRole)
      uiMessage.success('已分配角色')
      setAssignOpen(false)
      setSelectedRole('')
      adminRolesRequest.reload()
    } catch (err) {
      if (err instanceof Error) uiMessage.error(err.message)
    }
  }

  const handleRemoveRole = async (roleId: string, roleName: string) => {
    try {
      await platformApi.removeAdminRole(adminId, roleId)
      uiMessage.success(`已移除角色：${roleName}`)
      adminRolesRequest.reload()
    } catch (err) {
      if (err instanceof Error) uiMessage.error(err.message)
    }
  }

  return (
    <div className="page">
      <PageHeader
        title="平台管理员详情"
        description="查看管理员信息与角色"
        extra={
          <Space>
            <Button type="primary" onClick={() => setAssignOpen(true)}>
              分配角色
            </Button>
          </Space>
        }
      />
      <RequestBanner error={requestError} />
      <Card>
        {loading ? (
          <Skeleton active paragraph={{ rows: 3 }} />
        ) : !admin?.id ? (
          <Empty description="未找到该管理员" image={Empty.PRESENTED_IMAGE_SIMPLE} />
        ) : (
          <Descriptions column={1} bordered size="middle">
            <Descriptions.Item label="姓名">{admin.name}</Descriptions.Item>
            <Descriptions.Item label="邮箱">{admin.email || '-'}</Descriptions.Item>
            <Descriptions.Item label="状态">
              <Tag color={admin.status === 'active' ? 'green' : 'red'}>
                {admin.status === 'active' ? '启用' : '停用'}
              </Tag>
            </Descriptions.Item>
            <Descriptions.Item label="创建时间">{formatDateTime(admin.created_at)}</Descriptions.Item>
          </Descriptions>
        )}
      </Card>
      <Card>
        <TechnicalMeta items={[{ key: 'admin-id', label: 'Admin ID', value: admin.id || adminId }]} />
      </Card>
      <Card title={`已分配角色（${assignedRoles.length}）`}>
        {assignedRoles.length === 0 ? (
          <Empty description="该管理员暂未分配角色" image={Empty.PRESENTED_IMAGE_SIMPLE} />
        ) : (
          <Space wrap>
            {assignedRoles.map((role) => (
              <Space key={role.id} size={4}>
                <Tag color="blue">{role.name}</Tag>
                <Popconfirm
                  title={`移除角色「${role.name}」?`}
                  okText="移除"
                  cancelText="取消"
                  onConfirm={() => handleRemoveRole(role.id, role.name)}
                >
                  <Button type="link" danger size="small">
                    移除
                  </Button>
                </Popconfirm>
              </Space>
            ))}
          </Space>
        )}
        <Typography.Paragraph className="muted" style={{ marginTop: 12, marginBottom: 0 }}>
          可在右上角“分配角色”继续添加角色，或在此处移除已有角色。
        </Typography.Paragraph>
      </Card>

      <Modal
        title="分配角色"
        open={assignOpen}
        onCancel={() => setAssignOpen(false)}
        onOk={handleAssign}
        okText="保存"
      >
        <Select
          placeholder="选择角色"
          value={selectedRole || undefined}
          onChange={setSelectedRole}
          style={{ width: '100%' }}
          options={roleData.items
            .filter((role) => !assignedRoles.some((assigned) => assigned.id === role.id))
            .map((role) => ({ value: role.id, label: role.name }))}
        />
      </Modal>
    </div>
  )
}

