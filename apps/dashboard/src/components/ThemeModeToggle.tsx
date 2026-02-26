import { Switch, Tooltip } from 'antd'
import { BulbOutlined, MoonOutlined, SunOutlined } from '@ant-design/icons'
import { useThemeMode } from '../theme/mode'

export function ThemeModeToggle() {
  const { mode, setMode } = useThemeMode()
  const isDark = mode === 'dark'

  return (
    <Tooltip title="主题模式">
      <div className="theme-toggle">
        <SunOutlined className={`theme-toggle-icon ${!isDark ? 'is-active' : ''}`} />
        <Switch
          checked={isDark}
          onChange={(checked) => setMode(checked ? 'dark' : 'light')}
          aria-label="切换深浅色主题"
        />
        <MoonOutlined className={`theme-toggle-icon ${isDark ? 'is-active' : ''}`} />
      </div>
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
