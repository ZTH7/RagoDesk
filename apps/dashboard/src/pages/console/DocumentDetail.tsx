import { Button, Card, Descriptions, Form, Modal, Select, Space, Table, Tag, Popconfirm, message, Skeleton } from 'antd'
import { useMemo, useState } from 'react'
import { useParams } from 'react-router-dom'
import { PageHeader } from '../../components/PageHeader'
import { RequestBanner } from '../../components/RequestBanner'
import { useRequest } from '../../hooks/useRequest'
import { consoleApi } from '../../services/console'

export function DocumentDetail() {
  const { id } = useParams()
  const docId = id ?? ''
  const [rollbackOpen, setRollbackOpen] = useState(false)
  const [rollbackForm] = Form.useForm()

  const { data, loading, error, reload } = useRequest(
    () => consoleApi.getDocument(docId),
    {
      document: {
        id: '',
        title: '',
        source_type: '',
        status: '',
        current_version: 0,
        updated_at: '',
      },
      versions: [],
    },
    { enabled: Boolean(docId) },
  )

  const stageStatus = useMemo(() => {
    const status = data.document.status
    if (status === 'ready') return 'done'
    if (status === 'failed') return 'failed'
    if (status === 'processing') return 'processing'
    return 'unknown'
  }, [data.document.status])

  const timeline = useMemo(
    () =>
      ['uploaded', 'parsing', 'chunking', 'embedding', 'indexed'].map((stage) => ({
        stage,
        status: stageStatus,
      })),
    [stageStatus],
  )

  const handleReindex = async () => {
    try {
      await consoleApi.reindexDocument(docId)
      message.success('已触发 Reindex')
      reload()
    } catch (err) {
      if (err instanceof Error) message.error(err.message)
    }
  }

  const handleRollback = async () => {
    try {
      const values = await rollbackForm.validateFields()
      await consoleApi.rollbackDocument(docId, Number(values.version))
      message.success('已触发 Rollback')
      setRollbackOpen(false)
      reload()
    } catch (err) {
      if (err instanceof Error) message.error(err.message)
    }
  }

  return (
    <div className="page">
      <PageHeader
        title="文档详情"
        description="文档版本与 ingestion 状态"
        extra={
          <Space>
            <Button onClick={() => setRollbackOpen(true)}>Rollback</Button>
            <Popconfirm title="确认重新索引当前版本？" onConfirm={handleReindex}>
              <Button type="primary">Reindex</Button>
            </Popconfirm>
          </Space>
        }
      />
      <RequestBanner error={error} />
      <Card>
        {loading ? (
          <Skeleton active paragraph={{ rows: 3 }} />
        ) : (
          <Descriptions column={1} bordered size="middle">
            <Descriptions.Item label="Document ID">{data.document.id || docId}</Descriptions.Item>
            <Descriptions.Item label="标题">{data.document.title || '-'}</Descriptions.Item>
            <Descriptions.Item label="状态">
              <Tag color={data.document.status === 'ready' ? 'green' : 'gold'}>
                {data.document.status || 'unknown'}
              </Tag>
            </Descriptions.Item>
            <Descriptions.Item label="当前版本">v{data.document.current_version || '-'}</Descriptions.Item>
          </Descriptions>
        )}
      </Card>
      <Card title="版本历史">
        <Table
          rowKey="id"
          dataSource={data.versions}
          pagination={false}
          columns={[
            { title: 'Version', dataIndex: 'version' },
            { title: 'Status', dataIndex: 'status' },
            { title: 'Created At', dataIndex: 'created_at' },
          ]}
        />
      </Card>
      <Card title="Ingestion 时间线">
        <Table
          rowKey="stage"
          pagination={false}
          dataSource={timeline}
          columns={[
            { title: 'Stage', dataIndex: 'stage' },
            { title: 'Status', dataIndex: 'status' },
          ]}
        />
      </Card>

      <Modal
        title="回滚版本"
        open={rollbackOpen}
        onCancel={() => setRollbackOpen(false)}
        onOk={handleRollback}
        okText="确认回滚"
      >
        <Form form={rollbackForm} layout="vertical">
          <Form.Item label="回滚版本" name="version" rules={[{ required: true, message: '请选择版本' }]}>
            <Select
              placeholder="选择版本"
              options={data.versions.map((item) => ({
                value: item.version,
                label: `v${item.version} · ${item.status}`,
              }))}
            />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  )
}
