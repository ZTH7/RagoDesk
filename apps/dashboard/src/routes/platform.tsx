import { ApartmentOutlined, SafetyOutlined, TeamOutlined, LockOutlined } from '@ant-design/icons'
import { Tenants } from '../pages/platform/Tenants'
import { TenantDetail } from '../pages/platform/TenantDetail'
import { PlatformAdmins } from '../pages/platform/PlatformAdmins'
import { PlatformAdminDetail } from '../pages/platform/PlatformAdminDetail'
import { PlatformRoles } from '../pages/platform/PlatformRoles'
import { PlatformRoleDetail } from '../pages/platform/PlatformRoleDetail'
import { PlatformPermissions } from '../pages/platform/PlatformPermissions'
import type { AppRoute, NavItem } from './types'
import { permissions } from '../auth/permissions'

export const platformDefaultPath = '/platform/tenants'

export const platformNavItems: NavItem[] = [
  {
    key: '/platform/tenants',
    icon: <ApartmentOutlined />,
    label: '租户管理',
    permission: permissions.platform.tenantRead,
  },
  { key: '/platform/admins', icon: <TeamOutlined />, label: '平台管理员', permission: permissions.platform.adminRead },
  { key: '/platform/roles', icon: <SafetyOutlined />, label: '平台角色', permission: permissions.platform.roleRead },
  { key: '/platform/permissions', icon: <LockOutlined />, label: '权限目录', permission: permissions.platform.permissionRead },
]

export const platformMenuKeys = ['/platform/tenants', '/platform/admins', '/platform/roles', '/platform/permissions']

export const platformRoutes: AppRoute[] = [
  { path: 'tenants', element: <Tenants />, permission: permissions.platform.tenantRead },
  { path: 'tenants/:id', element: <TenantDetail />, permission: permissions.platform.tenantRead },
  { path: 'admins', element: <PlatformAdmins />, permission: permissions.platform.adminRead },
  { path: 'admins/:id', element: <PlatformAdminDetail />, permission: permissions.platform.adminRead },
  { path: 'roles', element: <PlatformRoles />, permission: permissions.platform.roleRead },
  { path: 'roles/:id', element: <PlatformRoleDetail />, permission: permissions.platform.roleRead },
  { path: 'permissions', element: <PlatformPermissions />, permission: permissions.platform.permissionRead },
]
