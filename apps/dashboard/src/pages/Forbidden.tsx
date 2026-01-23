import { Result } from 'antd'

export function Forbidden() {
  return (
    <Result
      status="403"
      title="没有权限"
      subTitle="你没有访问此页面的权限，请联系管理员。"
    />
  )
}
