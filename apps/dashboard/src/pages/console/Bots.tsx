import { Button, Form, Input, Modal, Popconfirm, Select, Space, Tag, message } from 'antd'
import { useMemo, useState } from 'react'
import { Link } from 'react-router-dom'
import { PageHeader } from '../../components/PageHeader'
import { FilterBar } from '../../components/FilterBar'
import { TableCard } from '../../components/TableCard'
import { DataSourceTag } from '../../components/DataSourceTag'
import { RequestBanner } from '../../components/RequestBanner'
import { useRequest } from '../../hooks/useRequest'
import { consoleApi } from '../../services/console'

const statusColors: Record<string, string> = {
  active: 'green',
  disabled: 'red',
}

export function Bots() {
  const [keyword, setKeyword] = useState('')
  const [statusFilter, setStatusFilter] = useState('all')
  const [modalOpen, setModalOpen] = useState(false)
  const [editingId, setEditingId] = useState<string | null>(null)
  const [form] = Form.useForm()
  const { data, loading, error, source, reload } = useRequest(() => consoleApi.listBots(), { items: [] })

  const filtered = useMemo(() => {
    return data.items.filter((item) => {
      if (statusFilter !== 'all' && item.status !== statusFilter) return false
      if (!keyword) return true
      return item.name.toLowerCase().includes(keyword.toLowerCase())
    })
  }, [data.items, keyword, statusFilter])

  const openCreate = () => {
    setEditingId(null)
    form.resetFields()
    setModalOpen(true)
  }

  const openEdit = (record: { id: string; name: string; description?: string; status: string }) => {
    setEditingId(record.id)
    form.setFieldsValue({ name: record.name, description: record.description, status: record.status })
    setModalOpen(true)
  }

  const handleSubmit = async () => {
    try {
      const values = await form.validateFields()
      if (editingId) {
        await consoleApi.updateBot(editingId, values)
        message.success('已更新机器人')
      } else {
        await consoleApi.createBot(values)
        message.success('已创建机器人')
      }
      setModalOpen(false)
      form.resetFields()
      reload()
    } catch (err) {
      if (err instanceof Error) message.error(err.message)
    }
  }

  const handleDelete = async (id: string) => {
    try {
      await consoleApi.deleteBot(id)
      message.success('已删除机器人')
      reload()
    } catch (err) {
      if (err instanceof Error) message.error(err.message)
    }
  }

  return (
    <div className="page">
      <PageHeader
        title="机器人"
        description="管理 Bot 与默认 RAG 流水线"
        extra={<DataSourceTag source={source} />}
      />
      <RequestBanner error={error} />
      <FilterBar
        left={<Input.Search placeholder="搜索 Bot" onSearch={setKeyword} allowClear style={{ width: 220 }} />}
        right={
          <Space>
            <Select
              defaultValue="all"
              style={{ width: 160 }}
              options={[
                { value: 'all', label: '全部状态' },
                { value: 'active', label: 'Active' },
                { value: 'disabled', label: 'Disabled' },
              ]}
              onChange={setStatusFilter}
            />
            <Button type="primary" onClick={openCreate}>
              新建机器人
            </Button>
          </Space>
        }
      />
      <TableCard
        table={{
          rowKey: 'id',
          dataSource: filtered,
          loading,
          pagination: { pageSize: 8 },
          columns: [
            {
              title: 'ID',
              dataIndex: 'id',
              render: (value: string) => <Link to={`/console/bots/${value}`}>{value}</Link>,
            },
            { title: '名称', dataIndex: 'name' },
            { title: '描述', dataIndex: 'description', render: (v: string) => v || '-' },
            {
              title: '状态',
              dataIndex: 'status',
              render: (value: string) => <Tag color={statusColors[value] || 'default'}>{value}</Tag>,
            },
            { title: '创建时间', dataIndex: 'created_at' },
            {
              title: '操作',
              key: 'actions',
              render: (_: unknown, record) => (
                <Space>
                  <Button size="small" onClick={() => openEdit(record)}>
                    编辑
                  </Button>
                  <Popconfirm title="确认删除该机器人？" onConfirm={() => handleDelete(record.id)}>
                    <Button size="small" danger>
                      删除
                    </Button>
                  </Popconfirm>
                </Space>
              ),
            },
          ],
        }}
      />

      <Modal
        title={editingId ? '编辑机器人' : '新建机器人'}
        open={modalOpen}
        onCancel={() => setModalOpen(false)}
        onOk={handleSubmit}
        okText={editingId ? '保存' : '创建'}
      >
        <Form form={form} layout="vertical" initialValues={{ status: 'active' }}>
          <Form.Item label="名称" name="name" rules={[{ required: true, message: '请输入名称' }]}>
            <Input placeholder="例如：产品客服 Bot" />
          </Form.Item>
          <Form.Item label="描述" name="description">
            <Input.TextArea placeholder="描述 Bot 用途" rows={3} />
          </Form.Item>
          <Form.Item label="状态" name="status" rules={[{ required: true }]}>
            <Select
              options={[
                { value: 'active', label: 'Active' },
                { value: 'disabled', label: 'Disabled' },
              ]}
            />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  )
}
