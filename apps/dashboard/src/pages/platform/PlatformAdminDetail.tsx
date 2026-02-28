import { Button, Card, Descriptions, Empty, Modal, Select, Space, Tag, Skeleton, Typography } from 'antd'
import { useState } from 'react'
import { useParams } from 'react-router-dom'
import { PageHeader } from '../../components/PageHeader'
import { TechnicalMeta } from '../../components/TechnicalMeta'
import { RequestBanner } from '../../components/RequestBanner'
import { useRequest } from '../../hooks/useRequest'
import { platformApi } from '../../services/platform'

import { uiMessage } from '../../services/uiMessage'
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
      uiMessage.error('请选择角色')
      return
    }
    try {
      await platformApi.assignAdminRole(adminId, selectedRole)
      uiMessage.success('已分配角色')
      setAssignOpen(false)
      setSelectedRole('')
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
      <RequestBanner error={error} />
      <Card>
        {loading ? (
          <Skeleton active paragraph={{ rows: 3 }} />
        ) : !admin ? (
          <Empty description="未找到该管理员" image={Empty.PRESENTED_IMAGE_SIMPLE} />
        ) : (
          <Descriptions column={1} bordered size="middle">
            <Descriptions.Item label="姓名">{admin.name}</Descriptions.Item>
            <Descriptions.Item label="状态">
              <Tag color={admin.status === 'active' ? 'green' : 'red'}>
                {admin.status === 'active' ? '启用' : '停用'}
              </Tag>
            </Descriptions.Item>
          </Descriptions>
        )}
      </Card>
      <Card>
        <TechnicalMeta items={[{ key: 'admin-id', label: 'Admin ID', value: admin?.id || adminId }]} />
      </Card>
      <Card title="可分配角色">
        {roleData.items.length === 0 ? (
          <Empty description="暂无可分配角色" image={Empty.PRESENTED_IMAGE_SIMPLE} />
        ) : (
          <Space wrap>
            {roleData.items.map((role) => (
              <Button
                key={role.id}
                size="small"
                onClick={async () => {
                  try {
                    await platformApi.assignAdminRole(adminId, role.id)
                    uiMessage.success(`已为管理员分配角色：${role.name}`)
                  } catch (err) {
                    if (err instanceof Error) uiMessage.error(err.message)
                  }
                }}
              >
                添加：{role.name}
              </Button>
            ))}
          </Space>
        )}
        <Typography.Paragraph className="muted" style={{ marginTop: 12, marginBottom: 0 }}>
          当前后端未提供“已分配角色列表”查询接口，此处提供快速分配入口。
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
          options={roleData.items.map((role) => ({ value: role.id, label: role.name }))}
        />
      </Modal>
    </div>
  )
}

