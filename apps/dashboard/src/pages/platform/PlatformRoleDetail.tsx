import { Card, Descriptions, Table, Tag, Space, Button, Modal, Select, Skeleton, Empty } from 'antd'
import { useState } from 'react'
import { useParams } from 'react-router-dom'
import { PageHeader } from '../../components/PageHeader'
import { TechnicalMeta } from '../../components/TechnicalMeta'
import { RequestBanner } from '../../components/RequestBanner'
import { useRequest } from '../../hooks/useRequest'
import { platformApi } from '../../services/platform'
import { formatDateTime } from '../../utils/datetime'

import { uiMessage } from '../../services/uiMessage'
export function PlatformRoleDetail() {
  const { id } = useParams()
  const roleId = id ?? ''
  const [editOpen, setEditOpen] = useState(false)
  const [permissionCodes, setPermissionCodes] = useState<string[]>([])

  const { data: roleData, loading: roleLoading, error: roleError } = useRequest(
    () => platformApi.getRole(roleId),
    { role: { id: '', name: '' } },
    { enabled: Boolean(roleId), deps: [roleId] },
  )
  const role = roleData.role

  const permRequest = useRequest(
    () => platformApi.listRolePermissions(roleId),
    { items: [] },
    { enabled: Boolean(roleId), deps: [roleId] },
  )
  const { data: permData } = permRequest
  const { data: allPerms } = useRequest(() => platformApi.listPermissions(), { items: [] })
  const requestError = roleError || permRequest.error

  const openEdit = () => {
    setPermissionCodes(permData.items.map((item) => item.code))
    setEditOpen(true)
  }

  const handleSave = async () => {
    try {
      await platformApi.assignRolePermissions(roleId, permissionCodes)
      uiMessage.success('已更新角色权限')
      setEditOpen(false)
      permRequest.reload()
    } catch (err) {
      if (err instanceof Error) uiMessage.error(err.message)
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
      <RequestBanner error={requestError} />
      <Card>
        {roleLoading ? (
          <Skeleton active paragraph={{ rows: 3 }} />
        ) : !role?.id ? (
          <Empty description="未找到该角色" image={Empty.PRESENTED_IMAGE_SIMPLE} />
        ) : (
          <Descriptions column={1} bordered size="middle">
            <Descriptions.Item label="名称">{role.name}</Descriptions.Item>
          </Descriptions>
        )}
      </Card>
      <Card>
        <TechnicalMeta items={[{ key: 'role-id', label: 'Role ID', value: role.id || roleId }]} />
      </Card>
      <Card title={`权限列表（${permData.items.length}）`}>
        <Table
          rowKey="code"
          dataSource={permData.items}
          pagination={false}
          columns={[
            { title: '权限标识', dataIndex: 'code' },
            {
              title: '权限域',
              dataIndex: 'scope',
              render: (value: string) => <Tag>{value === 'platform' ? '平台域' : value === 'tenant' ? '租户域' : value}</Tag>,
            },
            { title: '创建时间', dataIndex: 'created_at', render: (value: string) => formatDateTime(value) },
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

