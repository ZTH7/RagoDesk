import { Card, Typography } from 'antd'
import { PageHeader } from '../../components/PageHeader'

export function DevtoolsApi() {
  return (
    <div className="page">
      <PageHeader title="API 调试" description="使用 X-API-Key 调用外部接口" />
      <Card title="Quick Start">
        <Typography.Paragraph>
          1. 创建 Bot → 绑定 KB → 创建 API Key → 调用接口。
        </Typography.Paragraph>
        <div className="code-block">
          {`curl -X POST https://api.ragodesk.ai/api/v1/message \\
  -H "X-API-Key: <your_key>" \\
  -H "Content-Type: application/json" \\
  -d '{"session_id":"sess_abc","message":"如何申请退款？"}'`}
        </div>
      </Card>
    </div>
  )
}
