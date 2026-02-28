import { Collapse, Descriptions, Typography } from 'antd'

export type TechnicalMetaItem = {
  key: string
  label: string
  value: string | number | undefined | null
}

type TechnicalMetaProps = {
  items: TechnicalMetaItem[]
  title?: string
}

export function TechnicalMeta({ items, title = '技术信息（ID / 原始字段）' }: TechnicalMetaProps) {
  return (
    <Collapse
      size="small"
      items={[
        {
          key: 'technical-meta',
          label: title,
          children: (
            <Descriptions column={1} bordered size="small">
              {items.map((item) => (
                <Descriptions.Item key={item.key} label={item.label}>
                  <Typography.Text copyable={{ text: String(item.value ?? '-') }}>
                    {item.value ?? '-'}
                  </Typography.Text>
                </Descriptions.Item>
              ))}
            </Descriptions>
          ),
        },
      ]}
    />
  )
}

