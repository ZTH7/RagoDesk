import { Navigate, Route, Routes } from 'react-router-dom'
import { ConsoleLayout } from './layouts/ConsoleLayout'
import { PlatformLayout } from './layouts/PlatformLayout'
import { consoleRoutes, consoleDefaultPath } from './routes/console'
import { platformRoutes, platformDefaultPath } from './routes/platform'
import { NotFound } from './pages/NotFound'
import { ConsoleLogin } from './pages/auth/ConsoleLogin'
import { PlatformLogin } from './pages/auth/PlatformLogin'
import { RequirePermission } from './components/RequirePermission'
import { PermissionProvider } from './auth/PermissionContext'

function App() {
  return (
    <Routes>
      <Route path="/" element={<Navigate to={consoleDefaultPath} replace />} />
      <Route path="/console/login" element={<ConsoleLogin />} />
      <Route path="/platform/login" element={<PlatformLogin />} />

      <Route
        path="/console"
        element={
          <PermissionProvider scope="console">
            <ConsoleLayout />
          </PermissionProvider>
        }
      >
        <Route index element={<Navigate to="analytics/overview" replace />} />
        {consoleRoutes.map((route) => (
          <Route
            key={route.path}
            path={route.path}
            element={
              <RequirePermission permission={route.permission}>{route.element}</RequirePermission>
            }
          />
        ))}
      </Route>

      <Route
        path="/platform"
        element={
          <PermissionProvider scope="platform">
            <PlatformLayout />
          </PermissionProvider>
        }
      >
        <Route index element={<Navigate to={platformDefaultPath.replace('/platform/', '')} replace />} />
        {platformRoutes.map((route) => (
          <Route
            key={route.path}
            path={route.path}
            element={
              <RequirePermission permission={route.permission}>{route.element}</RequirePermission>
            }
          />
        ))}
      </Route>

      <Route path="*" element={<NotFound />} />
    </Routes>
  )
}

export default App
