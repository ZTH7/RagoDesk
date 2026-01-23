import { Button, Result } from 'antd'
import { useNavigate } from 'react-router-dom'

export function NotFound() {
  const navigate = useNavigate()
  return (
    <Result
      status="404"
      title="页面不存在"
      subTitle="请检查路径或返回首页。"
      extra={
        <Button type="primary" onClick={() => navigate('/console/analytics/overview')}>
          返回 Dashboard
        </Button>
      }
    />
  )
}
