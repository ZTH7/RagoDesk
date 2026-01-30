import { Button, Card, Descriptions, Form, Input, Modal, Space, Table, Tag, Popconfirm, message } from 'antd'
import { useState } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { PageHeader } from '../../components/PageHeader'
import { RequestBanner } from '../../components/RequestBanner'
import { useRequest } from '../../hooks/useRequest'
import { consoleApi } from '../../services/console'

export function KnowledgeBaseDetail() {
  const { id } = useParams()
  const navigate = useNavigate()
  const kbId = id ?? ''
  const [editOpen, setEditOpen] = useState(false)
  const [form] = Form.useForm()

  const { data: kbData, loading: kbLoading, error: kbError, reload } = useRequest(
    () => consoleApi.getKnowledgeBase(kbId),
    {
      knowledge_base: {
        id: '',
        name: '',
        description: '',
        created_at: '',
      },
    },
    { enabled: Boolean(kbId), deps: [kbId] },
  )

  const { data: docData, loading: docLoading } = useRequest(
    () => consoleApi.listDocuments({ kb_id: kbId }),
    { items: [] },
    { enabled: Boolean(kbId), deps: [kbId] },
  )

  const handleEdit = () => {
    form.setFieldsValue({
      name: kbData.knowledge_base.name,
      description: kbData.knowledge_base.description,
    })
    setEditOpen(true)
  }

  const handleSave = async () => {
    try {
      const values = await form.validateFields()
      await consoleApi.updateKnowledgeBase(kbId, values)
      message.success('已更新知识库')
      setEditOpen(false)
      reload()
    } catch (err) {
      if (err instanceof Error) message.error(err.message)
    }
  }

  const handleDelete = async () => {
    try {
      await consoleApi.deleteKnowledgeBase(kbId)
      message.success('已删除知识库')
      navigate('/console/knowledge-bases')
    } catch (err) {
      if (err instanceof Error) message.error(err.message)
    }
  }

  return (
    <div className="page">
      <PageHeader
        title="知识库详情"
        description="查看知识库信息与关联文档"
        extra={
          <Space>
            <Button onClick={handleEdit}>编辑</Button>
            <Popconfirm title="确认删除该知识库？" onConfirm={handleDelete}>
              <Button danger>删除</Button>
            </Popconfirm>
          </Space>
        }
      />
      <RequestBanner error={kbError} />
      <Card>
        {kbLoading ? (
          <Tag>Loading...</Tag>
        ) : (
          <Descriptions column={1} bordered size="middle">
            <Descriptions.Item label="KB ID">{kbData.knowledge_base.id || kbId}</Descriptions.Item>
            <Descriptions.Item label="名称">{kbData.knowledge_base.name || '-'}</Descriptions.Item>
            <Descriptions.Item label="描述">{kbData.knowledge_base.description || '-'}</Descriptions.Item>
          </Descriptions>
        )}
      </Card>
      <Card title="关联文档">
        <Table
          rowKey="id"
          dataSource={docData.items}
          loading={docLoading}
          pagination={false}
          columns={[
            { title: 'ID', dataIndex: 'id' },
            { title: '标题', dataIndex: 'title' },
            {
              title: '状态',
              dataIndex: 'status',
              render: (value: string) => <Tag color={value === 'ready' ? 'green' : 'gold'}>{value}</Tag>,
            },
            { title: '更新时间', dataIndex: 'updated_at' },
          ]}
        />
      </Card>

      <Modal
        title="编辑知识库"
        open={editOpen}
        onCancel={() => setEditOpen(false)}
        onOk={handleSave}
        okText="保存"
      >
        <Form form={form} layout="vertical">
          <Form.Item label="名称" name="name" rules={[{ required: true, message: '请输入名称' }]}>
            <Input />
          </Form.Item>
          <Form.Item label="描述" name="description">
            <Input.TextArea rows={3} />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  )
}
