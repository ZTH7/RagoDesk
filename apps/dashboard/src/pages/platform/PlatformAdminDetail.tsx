import { Alert, Button, Card, Descriptions, Modal, Select, Space, Tag, message, Skeleton } from 'antd'
import { useState } from 'react'
import { useParams } from 'react-router-dom'
import { PageHeader } from '../../components/PageHeader'
import { RequestBanner } from '../../components/RequestBanner'
import { useRequest } from '../../hooks/useRequest'
import { platformApi } from '../../services/platform'

export function PlatformAdminDetail() {
  const { id } = useParams()
  const adminId = id ?? ''
  const [assignOpen, setAssignOpen] = useState(false)
  const [selectedRole, setSelectedRole] = useState('')

  const { data, loading, error } = useRequest(() => platformApi.listAdmins(), { items: [] })
  const admin = data.items.find((item) => item.id === adminId)
  const { data: roleData } = useRequest(() => platformApi.listRoles(), { items: [] })

  const handleAssign = async () => {
    if (!selectedRole) {
      message.error('请选择角色')
      return
    }
    try {
      await platformApi.assignAdminRole(adminId, selectedRole)
      message.success('已分配角色')
      setAssignOpen(false)
      setSelectedRole('')
    } catch (err) {
      if (err instanceof Error) message.error(err.message)
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
      <RequestBanner error={error} />
      <Card>
        {loading ? (
          <Skeleton active paragraph={{ rows: 3 }} />
        ) : (
          <Descriptions column={1} bordered size="middle">
            <Descriptions.Item label="Admin ID">{admin?.id || adminId}</Descriptions.Item>
            <Descriptions.Item label="姓名">{admin?.name || '-'}</Descriptions.Item>
            <Descriptions.Item label="状态">
              <Tag color={admin?.status === 'active' ? 'green' : 'red'}>{admin?.status || 'unknown'}</Tag>
            </Descriptions.Item>
          </Descriptions>
        )}
      </Card>
      <Card title="角色列表">
        <Alert type="info" message="当前接口未提供管理员已分配角色列表" showIcon />
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
          options={roleData.items.map((role) => ({ value: role.id, label: role.name }))}
        />
      </Modal>
    </div>
  )
}
