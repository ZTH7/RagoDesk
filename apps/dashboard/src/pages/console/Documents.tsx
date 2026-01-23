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
  ready: 'green',
  processing: 'gold',
  failed: 'red',
}

const sourceOptions = [
  { value: 'pdf', label: 'PDF' },
  { value: 'docx', label: 'DOCX' },
  { value: 'markdown', label: 'Markdown' },
  { value: 'html', label: 'HTML' },
  { value: 'text', label: 'Text' },
  { value: 'url', label: 'URL' },
]

export function Documents() {
  const [status, setStatus] = useState<string>('all')
  const [keyword, setKeyword] = useState('')
  const [uploadOpen, setUploadOpen] = useState(false)
  const [form] = Form.useForm()
  const { data, loading, source, error, reload } = useRequest(() => consoleApi.listDocuments(), { items: [] })
  const { data: kbData } = useRequest(() => consoleApi.listKnowledgeBases(), { items: [] })

  const filtered = useMemo(() => {
    return data.items.filter((item) => {
      if (status !== 'all' && item.status !== status) return false
      if (keyword && !item.title.toLowerCase().includes(keyword.toLowerCase())) return false
      return true
    })
  }, [data.items, keyword, status])

  const handleUpload = async () => {
    try {
      const values = await form.validateFields()
      await consoleApi.uploadDocument(values)
      message.success('已提交文档上传')
      form.resetFields()
      setUploadOpen(false)
      reload()
    } catch (err) {
      if (err instanceof Error) message.error(err.message)
    }
  }

  const handleDelete = async (id: string) => {
    try {
      await consoleApi.deleteDocument(id)
      message.success('已删除文档')
      reload()
    } catch (err) {
      if (err instanceof Error) message.error(err.message)
    }
  }

  return (
    <div className="page">
      <PageHeader
        title="文档管理"
        description="上传、索引与管理文档版本"
        extra={<DataSourceTag source={source} />}
      />
      <RequestBanner error={error} />
      <FilterBar
        left={<Input.Search placeholder="搜索文档" onSearch={setKeyword} allowClear style={{ width: 220 }} />}
        right={
          <>
            <Select
              value={status}
              style={{ width: 160 }}
              onChange={setStatus}
              options={[
                { value: 'all', label: '全部状态' },
                { value: 'ready', label: 'Ready' },
                { value: 'processing', label: 'Processing' },
                { value: 'failed', label: 'Failed' },
              ]}
            />
            <Button type="primary" onClick={() => setUploadOpen(true)}>
              上传文档
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
              title: '标题',
              dataIndex: 'title',
              render: (_: string, record) => (
                <Link to={`/console/documents/${record.id}`}>{record.title}</Link>
              ),
            },
            { title: '类型', dataIndex: 'source_type' },
            {
              title: '状态',
              dataIndex: 'status',
              render: (value: string) => <Tag color={statusColors[value] || 'default'}>{value}</Tag>,
            },
            { title: '当前版本', dataIndex: 'current_version' },
            { title: '更新时间', dataIndex: 'updated_at' },
            {
              title: '操作',
              key: 'actions',
              render: (_: unknown, record) => (
                <Space>
                  <Popconfirm title="确认删除该文档？" onConfirm={() => handleDelete(record.id)}>
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
        title="上传文档"
        open={uploadOpen}
        onCancel={() => setUploadOpen(false)}
        onOk={handleUpload}
        okText="提交"
      >
        <Form form={form} layout="vertical">
          <Form.Item label="知识库" name="kb_id" rules={[{ required: true, message: '请选择知识库' }]}>
            <Select
              placeholder="选择知识库"
              options={kbData.items.map((kb) => ({
                value: kb.id,
                label: kb.name,
              }))}
              notFoundContent="暂无知识库，请先创建"
            />
          </Form.Item>
          <Form.Item label="标题" name="title" rules={[{ required: true, message: '请输入文档标题' }]}>
            <Input placeholder="例如：产品手册" />
          </Form.Item>
          <Form.Item label="来源类型" name="source_type" rules={[{ required: true, message: '请选择来源类型' }]}>
            <Select options={sourceOptions} />
          </Form.Item>
          <Form.Item label="Raw URI" name="raw_uri" rules={[{ required: true, message: '请输入对象存储地址' }]}>
            <Input placeholder="例如：oss://bucket/path/to/file.pdf" />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  )
}
