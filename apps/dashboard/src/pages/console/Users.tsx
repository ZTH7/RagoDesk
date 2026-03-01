import { Button, Descriptions, Form, Input, Modal, Select, Space, Switch, Tag, Typography } from 'antd'
import { useMemo, useState } from 'react'
import { PageHeader } from '../../components/PageHeader'
import { FilterBar } from '../../components/FilterBar'
import { TableCard } from '../../components/TableCard'
import { DataSourceTag } from '../../components/DataSourceTag'
import { RequestBanner } from '../../components/RequestBanner'
import { useRequest } from '../../hooks/useRequest'
import { consoleApi } from '../../services/console'
import { getCurrentTenantId } from '../../auth/storage'
import { formatDateTime } from '../../utils/datetime'

import { uiMessage } from '../../services/uiMessage'
const statusColors: Record<string, string> = {
  active: 'green',
  invited: 'orange',
  disabled: 'red',
}

const statusLabels: Record<string, string> = {
  active: '启用',
  invited: '待激活',
  disabled: '停用',
}

const normalizeAccount = (raw: string) => raw.trim()
const normalizePhone = (raw: string) => raw.replace(/[\s-]/g, '')
const looksLikeEmail = (value: string) => /\S+@\S+\.\S+/.test(value)
const looksLikePhone = (value: string) => /^\+?\d{6,20}$/.test(normalizePhone(value))

export function Users() {
  const tenantId = getCurrentTenantId() ?? ''
  const [status, setStatus] = useState<string>('all')
  const [keyword, setKeyword] = useState('')
  const [inviteOpen, setInviteOpen] = useState(false)
  const [assignOpen, setAssignOpen] = useState(false)
  const [assigningUserId, setAssigningUserId] = useState('')
  const [showAdvanced, setShowAdvanced] = useState(false)
  const [inviteLinkOpen, setInviteLinkOpen] = useState(false)
  const [inviteLink, setInviteLink] = useState('')
  const [inviteForm] = Form.useForm()
  const [assignForm] = Form.useForm()
  const sendInvite = Form.useWatch('send_invite', inviteForm) ?? true

  const { data, loading, source, error, reload } = useRequest(
    () => consoleApi.listUsers(tenantId),
    { items: [] },
    { enabled: Boolean(tenantId), deps: [tenantId] },
  )
  const { data: roleData } = useRequest(() => consoleApi.listRoles(), { items: [] })

  const filtered = useMemo(() => {
    return data.items.filter((item) => {
      if (status !== 'all' && item.status !== status) return false
      if (
        keyword &&
        !`${item.name || ''} ${item.email || ''} ${item.phone || ''}`
          .toLowerCase()
          .includes(keyword.toLowerCase())
      ) {
        return false
      }
      return true
    })
  }, [data.items, keyword, status])

  const handleInvite = async () => {
    if (!tenantId) {
      uiMessage.error('未获取到租户信息，请重新登录后重试')
      return
    }
    try {
      const values = await inviteForm.validateFields()
      const account = normalizeAccount(values.account)
      const res = await consoleApi.createUser(tenantId, {
        name: values.name,
        status: values.status,
        email: looksLikeEmail(account) ? account : undefined,
        phone: looksLikeEmail(account) ? undefined : normalizePhone(account),
        send_invite: values.send_invite,
        invite_base_url: values.send_invite ? values.invite_base_url?.trim() || window.location.origin : undefined,
        password: values.send_invite ? undefined : values.password,
      })
      uiMessage.success('已邀请成员')
      if (res.invite_link) {
        setInviteLink(res.invite_link)
        setInviteLinkOpen(true)
      }
      inviteForm.resetFields()
      setInviteOpen(false)
      reload()
    } catch (err) {
      if (err instanceof Error) uiMessage.error(err.message)
    }
  }

  const handleAssign = async () => {
    try {
      const values = await assignForm.validateFields()
      await consoleApi.assignRole(assigningUserId, values.role_id)
      uiMessage.success('已分配角色')
      setAssignOpen(false)
      assignForm.resetFields()
    } catch (err) {
      if (err instanceof Error) uiMessage.error(err.message)
    }
  }

  return (
    <div className="page">
      <PageHeader title="成员管理" description="邀请与管理租户成员" extra={<DataSourceTag source={source} />} />
      <RequestBanner error={error} />
      <FilterBar
        left={<Input.Search placeholder="按姓名或邮箱搜索" onSearch={setKeyword} allowClear style={{ width: 220 }} />}
        right={
          <>
            <Select
              value={status}
              style={{ width: 160 }}
              onChange={setStatus}
              options={[
                { value: 'all', label: '全部状态' },
                { value: 'active', label: '启用' },
                { value: 'invited', label: '待激活' },
                { value: 'disabled', label: '停用' },
              ]}
            />
            <Button type="primary" onClick={() => setInviteOpen(true)}>
              邀请成员
            </Button>
            <Space size={6}>
              <Typography.Text className="muted">高级列</Typography.Text>
              <Switch checked={showAdvanced} onChange={setShowAdvanced} />
            </Space>
          </>
        }
      />
      <TableCard
        table={{
          rowKey: 'id',
          dataSource: filtered,
          loading,
          pagination: { pageSize: 8 },
          expandable: showAdvanced
            ? {
                expandedRowRender: (record) => (
                  <Descriptions column={2} bordered size="small">
                    <Descriptions.Item label="成员 ID">{record.id}</Descriptions.Item>
                    <Descriptions.Item label="创建时间">{formatDateTime(record.created_at)}</Descriptions.Item>
                  </Descriptions>
                ),
              }
            : undefined,
          columns: [
            { title: '姓名', dataIndex: 'name' },
            { title: '邮箱', dataIndex: 'email' },
            {
              title: '状态',
              dataIndex: 'status',
              render: (value: string) => (
                <Tag color={statusColors[value] || 'default'}>{statusLabels[value] || value}</Tag>
              ),
            },
            { title: '创建时间', dataIndex: 'created_at', render: (value: string) => formatDateTime(value) },
            {
              title: '操作',
              key: 'actions',
              render: (_: unknown, record) => (
                <Space>
                  <Button
                    size="small"
                    onClick={() => {
                      setAssigningUserId(record.id)
                      setAssignOpen(true)
                    }}
                  >
                    分配角色
                  </Button>
                </Space>
              ),
            },
          ],
        }}
      />

      <Modal
        title="邀请成员"
        open={inviteOpen}
        onCancel={() => setInviteOpen(false)}
        onOk={handleInvite}
        okText="发送邀请"
      >
        <Form form={inviteForm} layout="vertical" initialValues={{ status: 'invited', send_invite: true }}>
          <Form.Item label="姓名" name="name" rules={[{ required: true, message: '请输入姓名' }]}>
            <Input placeholder="成员姓名" />
          </Form.Item>
          <Form.Item
            label="账号（邮箱/手机号）"
            name="account"
            rules={[
              { required: true, message: '请输入邮箱或手机号' },
              {
                validator: (_, value: string) => {
                  if (!value) return Promise.resolve()
                  const normalized = normalizeAccount(value)
                  if (looksLikeEmail(normalized) || looksLikePhone(normalized)) {
                    return Promise.resolve()
                  }
                  return Promise.reject(new Error('请输入合法邮箱或手机号'))
                },
              },
            ]}
            extra="支持邮箱或手机号，手机号中的空格与 - 会自动忽略"
          >
            <Input placeholder="name@company.com 或 +86 13800000000" allowClear />
          </Form.Item>
          <Form.Item label="发送邀请链接" name="send_invite" valuePropName="checked">
            <Switch />
          </Form.Item>
          {sendInvite ? (
            <Form.Item label="邀请链接基地址" name="invite_base_url" extra="默认使用当前站点地址">
              <Input placeholder="例如：http://localhost:5173" allowClear />
            </Form.Item>
          ) : (
            <Form.Item
              label="初始密码"
              name="password"
              rules={[
                { required: true, message: '请输入初始密码' },
                { min: 6, message: '密码至少 6 位' },
              ]}
            >
              <Input.Password placeholder="至少 6 位" />
            </Form.Item>
          )}
          <Form.Item label="状态" name="status">
            <Select
              options={[
                { value: 'active', label: '启用' },
                { value: 'invited', label: '待激活' },
                { value: 'disabled', label: '停用' },
              ]}
            />
          </Form.Item>
        </Form>
      </Modal>

      <Modal
        title="成员邀请链接"
        open={inviteLinkOpen}
        onCancel={() => setInviteLinkOpen(false)}
        footer={<Button onClick={() => setInviteLinkOpen(false)}>关闭</Button>}
      >
        <Typography.Text className="muted">请将该链接发送给成员完成首次登录。</Typography.Text>
        <Input.Group compact style={{ marginTop: 12 }}>
          <Input readOnly value={inviteLink} style={{ width: 'calc(100% - 84px)' }} />
          <Button
            onClick={async () => {
              try {
                await navigator.clipboard.writeText(inviteLink)
                uiMessage.success('邀请链接已复制')
              } catch {
                uiMessage.error('复制失败，请手动复制')
              }
            }}
          >
            复制
          </Button>
        </Input.Group>
      </Modal>

      <Modal
        title="分配角色"
        open={assignOpen}
        onCancel={() => setAssignOpen(false)}
        onOk={handleAssign}
        okText="保存"
      >
        <Form form={assignForm} layout="vertical">
          <Form.Item label="角色" name="role_id" rules={[{ required: true, message: '请选择角色' }]}>
            <Select
              placeholder="选择角色"
              options={roleData.items.map((role) => ({
                value: role.id,
                label: role.name,
              }))}
            />
          </Form.Item>
          {roleData.items.length === 0 ? (
            <Typography.Text className="muted">暂无角色，请先前往「角色管理」创建角色。</Typography.Text>
          ) : null}
        </Form>
      </Modal>
    </div>
  )
}

