import { Button, Form, Input, Modal, Popconfirm, Select, Space, Upload } from 'antd'
import type { UploadFile } from 'antd/es/upload/interface'
import { useMemo, useState } from 'react'
import { Link } from 'react-router-dom'
import { PageHeader } from '../../components/PageHeader'
import { FilterBar } from '../../components/FilterBar'
import { TableCard } from '../../components/TableCard'
import { DataSourceTag } from '../../components/DataSourceTag'
import { RequestBanner } from '../../components/RequestBanner'
import { useRequest } from '../../hooks/useRequest'
import { consoleApi } from '../../services/console'

import { uiMessage } from '../../services/uiMessage'
export function KnowledgeBases() {
  const [keyword, setKeyword] = useState('')
  const [modalOpen, setModalOpen] = useState(false)
  const [editingId, setEditingId] = useState<string | null>(null)
  const [uploadFiles, setUploadFiles] = useState<UploadFile[]>([])
  const [form] = Form.useForm()
  const { data, loading, source, error, reload } = useRequest(() => consoleApi.listKnowledgeBases(), { items: [] })

  const filtered = useMemo(() => {
    if (!keyword) return data.items
    return data.items.filter((item) => item.name.toLowerCase().includes(keyword.toLowerCase()))
  }, [data.items, keyword])

  const openCreate = () => {
    setEditingId(null)
    form.resetFields()
    setUploadFiles([])
    setModalOpen(true)
  }

  const openEdit = (record: { id: string; name: string; description: string }) => {
    setEditingId(record.id)
    form.setFieldsValue({ name: record.name, description: record.description })
    setModalOpen(true)
  }

  const handleSubmit = async () => {
    try {
      const values = await form.validateFields()
      if (editingId) {
        await consoleApi.updateKnowledgeBase(editingId, values)
        uiMessage.success('已更新知识库')
      } else {
        const created = await consoleApi.createKnowledgeBase(values)
        const kbId = created.knowledge_base?.id
        if (kbId && uploadFiles.length > 0) {
          const formData = new FormData()
          formData.append('kb_id', kbId)
          uploadFiles.forEach((file) => {
            if (file.originFileObj) {
              formData.append('files', file.originFileObj)
            }
          })
          await consoleApi.uploadDocumentFile(formData)
          uiMessage.success('已上传文档')
        }
        uiMessage.success('已创建知识库')
      }
      setModalOpen(false)
      form.resetFields()
      setUploadFiles([])
      reload()
    } catch (err) {
      if (err instanceof Error) uiMessage.error(err.message)
    }
  }

  const handleDelete = async (id: string) => {
    try {
      await consoleApi.deleteKnowledgeBase(id)
      uiMessage.success('已删除知识库')
      reload()
    } catch (err) {
      if (err instanceof Error) uiMessage.error(err.message)
    }
  }

  return (
    <div className="page">
      <PageHeader
        title="知识库"
        description="管理知识库与文档集合"
        extra={<DataSourceTag source={source} />}
      />
      <RequestBanner error={error} />
      <FilterBar
        left={<Input.Search placeholder="搜索知识库" onSearch={setKeyword} allowClear style={{ width: 220 }} />}
        right={
          <>
            <Select defaultValue="all" style={{ width: 160 }} options={[{ value: 'all', label: '全部类型' }]} />
            <Button type="primary" onClick={openCreate}>
              新建知识库
            </Button>
          </>
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
              render: (value: string) => <Link to={`/console/knowledge-bases/${value}`}>{value}</Link>,
            },
            { title: '名称', dataIndex: 'name' },
            { title: '描述', dataIndex: 'description' },
            { title: '文档数量', dataIndex: 'document_count', render: (v) => v ?? '-' },
            { title: '创建时间', dataIndex: 'created_at' },
            {
              title: '操作',
              key: 'actions',
              render: (_: unknown, record) => (
                <Space>
                  <Button size="small" onClick={() => openEdit(record)}>
                    编辑
                  </Button>
                  <Popconfirm title="确认删除该知识库？" onConfirm={() => handleDelete(record.id)}>
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
        title={editingId ? '编辑知识库' : '新建知识库'}
        open={modalOpen}
        onCancel={() => setModalOpen(false)}
        onOk={handleSubmit}
        okText={editingId ? '保存' : '创建'}
      >
        <Form form={form} layout="vertical">
          <Form.Item label="名称" name="name" rules={[{ required: true, message: '请输入名称' }]}>
            <Input placeholder="例如：产品知识库" />
          </Form.Item>
          <Form.Item label="描述" name="description">
            <Input.TextArea placeholder="简要说明知识库用途" rows={3} />
          </Form.Item>
          {!editingId && (
            <Form.Item label="上传文档">
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
            </Form.Item>
          )}
        </Form>
      </Modal>
    </div>
  )
}

