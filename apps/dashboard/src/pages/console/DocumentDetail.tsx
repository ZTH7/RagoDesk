import {
  Button,
  Card,
  Descriptions,
  Form,
  Modal,
  Select,
  Space,
  Table,
  Tag,
  Popconfirm,
  Skeleton,
} from 'antd'
import { useMemo, useState } from 'react'
import { useParams } from 'react-router-dom'
import { PageHeader } from '../../components/PageHeader'
import { RequestBanner } from '../../components/RequestBanner'
import { useRequest } from '../../hooks/useRequest'
import { consoleApi } from '../../services/console'

import { uiMessage } from '../../services/uiMessage'
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
    { enabled: Boolean(docId), deps: [docId] },
  )

  const timeline = useMemo(() => {
    const stages = ['uploaded', 'parsing', 'chunking', 'embedding', 'indexed']
    const status = data.document.status
    return stages.map((stage, idx) => {
      if (status === 'ready') {
        return { stage, status: 'done' }
      }
      if (status === 'failed') {
        return { stage, status: idx === 0 ? 'done' : 'failed' }
      }
      if (status === 'processing') {
        return { stage, status: idx === 0 ? 'done' : 'processing' }
      }
      if (status === 'uploaded') {
        return { stage, status: idx === 0 ? 'done' : 'pending' }
      }
      return { stage, status: idx === 0 ? 'done' : 'pending' }
    })
  }, [data.document.status])

  const renderStageStatus = (value: string) => {
    const color =
      value === 'done'
        ? 'green'
        : value === 'failed'
          ? 'red'
          : value === 'processing'
            ? 'gold'
            : 'default'
    return <Tag color={color}>{value}</Tag>
  }

  const handleReindex = async () => {
    try {
      await consoleApi.reindexDocument(docId)
      uiMessage.success('已触发 Reindex')
      reload()
    } catch (err) {
      if (err instanceof Error) uiMessage.error(err.message)
    }
  }

  const handleRollback = async () => {
    try {
      const values = await rollbackForm.validateFields()
      await consoleApi.rollbackDocument(docId, Number(values.version))
      uiMessage.success('已触发 Rollback')
      setRollbackOpen(false)
      reload()
    } catch (err) {
      if (err instanceof Error) uiMessage.error(err.message)
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
            <Descriptions.Item label="来源类型">{data.document.source_type || '-'}</Descriptions.Item>
            <Descriptions.Item label="最近更新">{data.document.updated_at || '-'}</Descriptions.Item>
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
            { title: 'Status', dataIndex: 'status', render: renderStageStatus },
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

