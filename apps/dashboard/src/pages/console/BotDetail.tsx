import {
  Button,
  Card,
  Descriptions,
  Form,
  Input,
  InputNumber,
  Modal,
  Popconfirm,
  Select,
  Space,
  Table,
  Tag,
  Skeleton,
} from 'antd'
import { useState } from 'react'
import { useParams } from 'react-router-dom'
import { PageHeader } from '../../components/PageHeader'
import { RequestBanner } from '../../components/RequestBanner'
import { useRequest } from '../../hooks/useRequest'
import { consoleApi } from '../../services/console'

import { uiMessage } from '../../services/uiMessage'
export function BotDetail() {
  const { id } = useParams()
  const botId = id ?? ''
  const [bindOpen, setBindOpen] = useState(false)
  const [editOpen, setEditOpen] = useState(false)
  const [bindForm] = Form.useForm()
  const [editForm] = Form.useForm()

  const {
    data: botData,
    loading: botLoading,
    error: botError,
    reload: reloadBot,
  } = useRequest(
    () => consoleApi.getBot(botId),
    { bot: { id: '', name: '', status: '', created_at: '' } },
    { enabled: Boolean(botId), deps: [botId] },
  )
  const bot = botData.bot

  const { data: kbData, reload: reloadBindings } = useRequest(
    () => consoleApi.listBotKnowledgeBases(botId),
    { items: [] },
    { enabled: Boolean(botId), deps: [botId] },
  )
  const { data: keyData } = useRequest(
    () => consoleApi.listApiKeys({ bot_id: botId }),
    { items: [] },
    { enabled: Boolean(botId), deps: [botId] },
  )
  const { data: allKBs } = useRequest(() => consoleApi.listKnowledgeBases(), { items: [] })

  const handleBind = async () => {
    try {
      const values = await bindForm.validateFields()
      await consoleApi.bindBotKnowledgeBase(botId, values.kb_id, values.weight)
      uiMessage.success('已绑定知识库')
      setBindOpen(false)
      bindForm.resetFields()
      reloadBindings()
    } catch (err) {
      if (err instanceof Error) uiMessage.error(err.message)
    }
  }

  const handleUnbind = async (kbId: string) => {
    try {
      await consoleApi.unbindBotKnowledgeBase(botId, kbId)
      uiMessage.success('已解绑知识库')
      reloadBindings()
    } catch (err) {
      if (err instanceof Error) uiMessage.error(err.message)
    }
  }

  const openEdit = () => {
    if (!bot) return
    editForm.setFieldsValue({
      name: bot.name,
      description: bot.description,
      status: bot.status,
    })
    setEditOpen(true)
  }

  const handleEdit = async () => {
    try {
      const values = await editForm.validateFields()
      await consoleApi.updateBot(botId, values)
      uiMessage.success('已更新机器人')
      setEditOpen(false)
      reloadBot()
    } catch (err) {
      if (err instanceof Error) uiMessage.error(err.message)
    }
  }

  return (
    <div className="page">
      <PageHeader
        title="机器人详情"
        description="查看机器人配置与关联资源"
        extra={
          <Space>
            <Button onClick={openEdit}>编辑</Button>
            <Button type="primary" onClick={() => setBindOpen(true)}>
              绑定知识库
            </Button>
          </Space>
        }
      />
      <RequestBanner error={botError} />
      <Card>
        {botLoading ? (
          <Skeleton active paragraph={{ rows: 3 }} />
        ) : (
          <Descriptions column={1} bordered size="middle">
            <Descriptions.Item label="Bot ID">{bot?.id || botId}</Descriptions.Item>
            <Descriptions.Item label="状态">
              <Tag color={bot?.status === 'active' ? 'green' : 'red'}>{bot?.status || 'unknown'}</Tag>
            </Descriptions.Item>
            <Descriptions.Item label="名称">{bot?.name || '-'}</Descriptions.Item>
            <Descriptions.Item label="描述">{bot?.description || '-'}</Descriptions.Item>
            <Descriptions.Item label="RAG Pipeline">标准工作流 + rerank</Descriptions.Item>
          </Descriptions>
        )}
      </Card>
      <Card title="关联知识库">
        <Table
          rowKey="id"
          dataSource={kbData.items}
          pagination={false}
          columns={[
            { title: 'ID', dataIndex: 'id' },
            { title: 'KB ID', dataIndex: 'kb_id' },
            { title: 'Weight', dataIndex: 'weight' },
            {
              title: '操作',
              key: 'actions',
              render: (_: unknown, record) => (
                <Popconfirm title="确认解绑该知识库？" onConfirm={() => handleUnbind(record.kb_id)}>
                  <Button size="small" danger>
                    解绑
                  </Button>
                </Popconfirm>
              ),
            },
          ]}
        />
      </Card>
      <Card title="关联 API Keys">
        <Table
          rowKey="id"
          dataSource={keyData.items}
          pagination={false}
          columns={[
            { title: '名称', dataIndex: 'name' },
            { title: '状态', dataIndex: 'status', render: (v) => <Tag color="green">{v}</Tag> },
            { title: 'Scopes', dataIndex: 'scopes', render: (v) => v.join(', ') },
          ]}
        />
      </Card>

      <Modal
        title="绑定知识库"
        open={bindOpen}
        onCancel={() => setBindOpen(false)}
        onOk={handleBind}
        okText="绑定"
      >
        <Form form={bindForm} layout="vertical">
          <Form.Item label="知识库" name="kb_id" rules={[{ required: true, message: '请选择知识库' }]}>
            <Select
              options={allKBs.items.map((kb) => ({ value: kb.id, label: kb.name }))}
              placeholder="选择知识库"
            />
          </Form.Item>
          <Form.Item label="权重" name="weight">
            <InputNumber min={0} max={1} step={0.1} style={{ width: '100%' }} />
          </Form.Item>
        </Form>
      </Modal>

      <Modal
        title="编辑机器人"
        open={editOpen}
        onCancel={() => setEditOpen(false)}
        onOk={handleEdit}
        okText="保存"
      >
        <Form form={editForm} layout="vertical">
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

