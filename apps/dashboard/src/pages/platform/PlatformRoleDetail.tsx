import { Card, Descriptions, Table, Tag, Space, Button, Modal, Select, message, Skeleton } from 'antd'
import { useState } from 'react'
import { useParams } from 'react-router-dom'
import { PageHeader } from '../../components/PageHeader'
import { RequestBanner } from '../../components/RequestBanner'
import { useRequest } from '../../hooks/useRequest'
import { platformApi } from '../../services/platform'

export function PlatformRoleDetail() {
  const { id } = useParams()
  const roleId = id ?? ''
  const [editOpen, setEditOpen] = useState(false)
  const [permissionCodes, setPermissionCodes] = useState<string[]>([])

  const { data: roleData, loading: roleLoading, error: roleError } = useRequest(
    () => platformApi.listRoles(),
    { items: [] },
  )
  const role = roleData.items.find((item) => item.id === roleId)

  const permRequest = useRequest(
    () => platformApi.listRolePermissions(roleId),
    { items: [] },
    { enabled: Boolean(roleId), deps: [roleId] },
  )
  const { data: permData } = permRequest
  const { data: allPerms } = useRequest(() => platformApi.listPermissions(), { items: [] })

  const openEdit = () => {
    setPermissionCodes(permData.items.map((item) => item.code))
    setEditOpen(true)
  }

  const handleSave = async () => {
    try {
      await platformApi.assignRolePermissions(roleId, permissionCodes)
      message.success('已更新角色权限')
      setEditOpen(false)
      permRequest.reload()
    } catch (err) {
      if (err instanceof Error) message.error(err.message)
    }
  }

  return (
    <div className="page">
      <PageHeader
        title="平台角色详情"
        description="角色权限与授权记录"
        extra={
          <Space>
            <Button onClick={openEdit}>编辑权限</Button>
          </Space>
        }
      />
      <RequestBanner error={roleError} />
      <Card>
        {roleLoading ? (
          <Skeleton active paragraph={{ rows: 3 }} />
        ) : (
          <Descriptions column={1} bordered size="middle">
            <Descriptions.Item label="Role ID">{role?.id || roleId}</Descriptions.Item>
            <Descriptions.Item label="名称">{role?.name || '-'}</Descriptions.Item>
          </Descriptions>
        )}
      </Card>
      <Card title="权限列表">
        <Table
          rowKey="code"
          dataSource={permData.items}
          pagination={false}
          columns={[
            { title: 'Code', dataIndex: 'code' },
            { title: 'Scope', dataIndex: 'scope', render: (value: string) => <Tag>{value}</Tag> },
            { title: '创建时间', dataIndex: 'created_at' },
          ]}
        />
      </Card>

      <Modal
        title="编辑角色权限"
        open={editOpen}
        onCancel={() => setEditOpen(false)}
        onOk={handleSave}
        okText="保存"
      >
        <Select
          mode="multiple"
          value={permissionCodes}
          onChange={setPermissionCodes}
          style={{ width: '100%' }}
          options={allPerms.items.map((item) => ({
            value: item.code,
            label: item.description ? `${item.code} · ${item.description}` : item.code,
          }))}
        />
      </Modal>
    </div>
  )
}
