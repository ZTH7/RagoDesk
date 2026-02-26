import { theme as antdTheme } from 'antd'
import type { ThemeConfig } from 'antd'

export type ResolvedThemeMode = 'light' | 'dark'

export function createTheme(mode: ResolvedThemeMode): ThemeConfig {
  const isDark = mode === 'dark'
  return {
    algorithm: isDark ? antdTheme.darkAlgorithm : antdTheme.defaultAlgorithm,
    token: {
      colorPrimary: '#1B4B66',
      colorInfo: '#2BB3B1',
      colorSuccess: '#16A34A',
      colorWarning: '#D97706',
      colorError: '#DC2626',
      fontFamily: '"IBM Plex Sans", -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif',
      fontFamilyCode: '"IBM Plex Mono", ui-monospace, SFMono-Regular, Menlo, monospace',
      borderRadius: 10,
      borderRadiusLG: 14,
      borderRadiusSM: 8,
      motionDurationFast: '0.18s',
      motionDurationMid: '0.26s',
      motionDurationSlow: '0.34s',
    },
    components: {
      Layout: {
        headerBg: isDark ? '#0f172a' : '#ffffff',
        bodyBg: isDark ? '#0b1220' : '#f6f7fb',
      },
      Menu: {
        itemBorderRadius: 10,
        itemHeight: 42,
      },
      Card: {
        borderRadiusLG: 14,
      },
      Button: {
        borderRadius: 10,
      },
      Input: {
        borderRadius: 10,
      },
      Select: {
        borderRadius: 10,
      },
      Table: {
        borderRadiusLG: 12,
      },
    },
  }
}
