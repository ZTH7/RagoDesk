import { Alert, Card, Descriptions, Skeleton } from 'antd'
import { useParams } from 'react-router-dom'
import { PageHeader } from '../../components/PageHeader'
import { TechnicalMeta } from '../../components/TechnicalMeta'
import { RequestBanner } from '../../components/RequestBanner'
import { useRequest } from '../../hooks/useRequest'
import { platformApi } from '../../services/platform'
import { formatDateTime } from '../../utils/datetime'

export function TenantDetail() {
  const { id } = useParams()
  const tenantId = id ?? ''
  const { data, loading, error } = useRequest(
    () => platformApi.getTenant(tenantId),
    {
      tenant: {
        id: '',
        name: '',
        type: '',
        plan: '',
        status: '',
        created_at: '',
      },
    },
    { enabled: Boolean(tenantId), deps: [tenantId] },
  )

  return (
    <div className="page">
      <PageHeader title="租户详情" description="查看租户概览与资源使用" />
      <RequestBanner error={error} />
      <Card>
        {loading ? (
          <Skeleton active paragraph={{ rows: 3 }} />
        ) : (
          <Descriptions column={1} bordered size="middle">
            <Descriptions.Item label="名称">{data.tenant.name || '-'}</Descriptions.Item>
            <Descriptions.Item label="类型">
              {data.tenant.type === 'enterprise' ? '企业' : data.tenant.type === 'personal' ? '个人' : data.tenant.type || '-'}
            </Descriptions.Item>
            <Descriptions.Item label="套餐">{data.tenant.plan || '-'}</Descriptions.Item>
            <Descriptions.Item label="状态">
              {data.tenant.status === 'active' ? '启用' : data.tenant.status === 'suspended' ? '停用' : data.tenant.status || '-'}
            </Descriptions.Item>
            <Descriptions.Item label="创建时间">{formatDateTime(data.tenant.created_at)}</Descriptions.Item>
          </Descriptions>
        )}
      </Card>
      <Card>
        <TechnicalMeta items={[{ key: 'tenant-id', label: 'Tenant ID', value: data.tenant.id || tenantId }]} />
      </Card>
      <Card>
        <Alert type="info" title="租户资源统计能力建设中，当前先提供基础资料管理。" showIcon />
      </Card>
    </div>
  )
}
