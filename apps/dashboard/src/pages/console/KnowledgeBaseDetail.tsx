import {
  Button,
  Card,
  Descriptions,
  Form,
  Input,
  Modal,
  Popconfirm,
  Select,
  Space,
  Table,
  Tag,
  Upload,
} from 'antd'
import type { UploadFile } from 'antd/es/upload/interface'
import { useMemo, useState } from 'react'
import { Link, useParams, useNavigate } from 'react-router-dom'
import { PageHeader } from '../../components/PageHeader'
import { TechnicalMeta } from '../../components/TechnicalMeta'
import { RequestBanner } from '../../components/RequestBanner'
import { useRequest } from '../../hooks/useRequest'
import { consoleApi } from '../../services/console'

import { uiMessage } from '../../services/uiMessage'
export function KnowledgeBaseDetail() {
  const { id } = useParams()
  const navigate = useNavigate()
  const kbId = id ?? ''
  const [editOpen, setEditOpen] = useState(false)
  const [uploadFiles, setUploadFiles] = useState<UploadFile[]>([])
  const [bindOpen, setBindOpen] = useState(false)
  const [bindDocIds, setBindDocIds] = useState<string[]>([])
  const [form] = Form.useForm()

  const { data: kbData, loading: kbLoading, error: kbError, reload: reloadKB } = useRequest(
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

  const { data: docData, loading: docLoading, reload: reloadDocs } = useRequest(
    () => consoleApi.listDocuments({ kb_id: kbId }),
    { items: [] },
    { enabled: Boolean(kbId), deps: [kbId] },
  )

  const { data: allDocsData, loading: allDocsLoading, reload: reloadAllDocs } = useRequest(
    () => consoleApi.listDocuments(),
    { items: [] },
  )

  const { data: kbListData } = useRequest(() => consoleApi.listKnowledgeBases(), { items: [] })

  const kbNameMap = useMemo(() => {
    return new Map(kbListData.items.map((kb) => [kb.id, kb.name]))
  }, [kbListData.items])

  const bindCandidates = useMemo(() => {
    return allDocsData.items.filter((doc) => doc.id && doc.kb_id !== kbId)
  }, [allDocsData.items, kbId])

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
      uiMessage.success('已更新知识库')
      setEditOpen(false)
      reloadKB()
    } catch (err) {
      if (err instanceof Error) uiMessage.error(err.message)
    }
  }

  const handleDelete = async () => {
    try {
      await consoleApi.deleteKnowledgeBase(kbId)
      uiMessage.success('已删除知识库')
      navigate('/console/knowledge-bases')
    } catch (err) {
      if (err instanceof Error) uiMessage.error(err.message)
    }
  }

  const handleUpload = async () => {
    if (!kbId) return
    if (uploadFiles.length === 0) {
      uiMessage.warning('请先选择文件')
      return
    }
    try {
      const formData = new FormData()
      formData.append('kb_id', kbId)
      uploadFiles.forEach((file) => {
        if (file.originFileObj) {
          formData.append('files', file.originFileObj)
        }
      })
      await consoleApi.uploadDocumentFile(formData)
      uiMessage.success('已提交文档上传')
      setUploadFiles([])
      reloadKB()
      reloadDocs()
    } catch (err) {
      if (err instanceof Error) uiMessage.error(err.message)
    }
  }

  const handleUnbind = async (documentId: string) => {
    if (!kbId) return
    try {
      await consoleApi.updateDocument(documentId, { kb_id: '' })
      uiMessage.success('已解绑文档')
      reloadDocs()
      reloadAllDocs()
    } catch (err) {
      if (err instanceof Error) uiMessage.error(err.message)
    }
  }

  const handleBind = async () => {
    if (!kbId) return
    if (bindDocIds.length === 0) {
      uiMessage.warning('请选择要绑定的文档')
      return
    }
    try {
      await Promise.all(bindDocIds.map((docId) => consoleApi.updateDocument(docId, { kb_id: kbId })))
      uiMessage.success('已绑定文档')
      setBindDocIds([])
      setBindOpen(false)
      reloadDocs()
      reloadAllDocs()
    } catch (err) {
      if (err instanceof Error) uiMessage.error(err.message)
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
            <Descriptions.Item label="名称">{kbData.knowledge_base.name || '-'}</Descriptions.Item>
            <Descriptions.Item label="描述">{kbData.knowledge_base.description || '-'}</Descriptions.Item>
          </Descriptions>
        )}
      </Card>
      <Card>
        <TechnicalMeta
          items={[
            { key: 'kb-id', label: 'Knowledge Base ID', value: kbData.knowledge_base.id || kbId },
          ]}
        />
      </Card>
      <Card
        title="关联文档"
        extra={
          <Button onClick={() => setBindOpen(true)} disabled={!kbId}>
            绑定文档
          </Button>
        }
      >
        <Table
          rowKey="id"
          dataSource={docData.items}
          loading={docLoading}
          pagination={false}
          columns={[
            {
              title: '标题',
              dataIndex: 'title',
              render: (_: string, record) => <Link to={`/console/documents/${record.id}`}>{record.title}</Link>,
            },
            {
              title: '状态',
              dataIndex: 'status',
              render: (value: string) => <Tag color={value === 'ready' ? 'green' : 'gold'}>{value}</Tag>,
            },
            { title: '更新时间', dataIndex: 'updated_at' },
            {
              title: '操作',
              key: 'actions',
              render: (_: unknown, record) => (
                <Popconfirm title="确认解绑该文档？" onConfirm={() => handleUnbind(record.id)}>
                  <Button size="small">解绑</Button>
                </Popconfirm>
              ),
            },
          ]}
        />
      </Card>
      <Card title="上传文档" style={{ marginTop: 16 }}>
        <Upload.Dragger
          multiple
          fileList={uploadFiles}
          beforeUpload={() => false}
          onChange={(info) => setUploadFiles(info.fileList)}
        >
          <p className="ant-upload-drag-icon">文件</p>
          <p className="ant-upload-text">点击或拖拽上传多个文件</p>
          <p className="ant-upload-hint">支持 PDF / DOCX / Markdown / HTML / 文本</p>
        </Upload.Dragger>
        <div style={{ marginTop: 12 }}>
          <Button type="primary" onClick={handleUpload} disabled={!kbId}>
            开始上传
          </Button>
        </div>
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

      <Modal
        title="绑定文档"
        open={bindOpen}
        onCancel={() => setBindOpen(false)}
        onOk={handleBind}
        okText="绑定"
      >
        <Form layout="vertical">
          <Form.Item label="选择文档">
            <Select
              mode="multiple"
              placeholder="选择要绑定到当前知识库的文档"
              value={bindDocIds}
              onChange={(value) => setBindDocIds(value)}
              loading={allDocsLoading}
              options={bindCandidates.map((doc) => ({
                value: doc.id,
                label: `${doc.title}${
                  doc.kb_id
                    ? ` · 当前：${kbNameMap.get(doc.kb_id) ?? '其他知识库'}`
                    : ' · 未绑定'
                }`,
              }))}
              notFoundContent="暂无可绑定文档"
            />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  )
}

