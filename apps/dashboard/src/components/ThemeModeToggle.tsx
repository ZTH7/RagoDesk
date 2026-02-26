import { Segmented, Tooltip } from 'antd'
import { BulbOutlined, DesktopOutlined, MoonOutlined, SunOutlined } from '@ant-design/icons'
import { useThemeMode } from '../theme/mode'

export function ThemeModeToggle() {
  const { mode, setMode } = useThemeMode()

  return (
    <Tooltip title="主题模式">
      <Segmented
        size="middle"
        value={mode}
        onChange={(value) => setMode(value as 'light' | 'dark' | 'system')}
        options={[
          { value: 'light', icon: <SunOutlined />, label: '浅色' },
          { value: 'dark', icon: <MoonOutlined />, label: '深色' },
          { value: 'system', icon: <DesktopOutlined />, label: '跟随系统' },
        ]}
      />
    </Tooltip>
  )
}

export function ThemeStatusDot() {
  const { resolvedMode } = useThemeMode()
  return (
    <span className={`theme-dot ${resolvedMode === 'dark' ? 'is-dark' : 'is-light'}`}>
      <BulbOutlined />
    </span>
  )
}
