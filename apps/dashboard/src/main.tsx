import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import { BrowserRouter } from 'react-router-dom'
import { App as AntdApp, ConfigProvider } from 'antd'
import 'antd/dist/reset.css'
import './index.css'
import App from './App'
import { ThemeModeProvider, useThemeMode } from './theme/mode'
import { createTheme } from './theme/theme'

function ThemedApp() {
  const { resolvedMode } = useThemeMode()
  return (
    <ConfigProvider theme={createTheme(resolvedMode)}>
      <AntdApp>
        <BrowserRouter>
          <App />
        </BrowserRouter>
      </AntdApp>
    </ConfigProvider>
  )
}

createRoot(document.getElementById('root')!).render(
  <StrictMode>
    <ThemeModeProvider>
      <ThemedApp />
    </ThemeModeProvider>
  </StrictMode>,
)
