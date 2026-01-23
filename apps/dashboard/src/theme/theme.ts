import type { ThemeConfig } from 'antd'

export const theme: ThemeConfig = {
  token: {
    colorPrimary: '#1B4B66',
    colorInfo: '#2BB3B1',
    colorSuccess: '#16A34A',
    colorWarning: '#D97706',
    colorError: '#DC2626',
    fontFamily: '"IBM Plex Sans", -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif',
    fontFamilyCode: '"IBM Plex Mono", ui-monospace, SFMono-Regular, Menlo, monospace',
    borderRadius: 6,
  },
  components: {
    Layout: {
      headerBg: '#FFFFFF',
      bodyBg: '#F6F7FB',
    },
    Menu: {
      itemBorderRadius: 6,
      itemHeight: 40,
    },
    Card: {
      borderRadiusLG: 10,
    },
  },
}
