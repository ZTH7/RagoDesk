import type { MenuProps } from 'antd'
import type { NavItem } from './types'

export function resolveSelectedKey(pathname: string, menuKeys: string[]) {
  const match = menuKeys
    .filter((key) => pathname.startsWith(key))
    .sort((a, b) => b.length - a.length)[0]
  return match ?? menuKeys[0]
}

export function buildMenuItems(items: NavItem[], permissions: Set<string>): MenuProps['items'] {
  const filtered = filterNavItems(items, permissions)
  return filtered.map((item) => toMenuItem(item, permissions))
}

function filterNavItems(items: NavItem[], permissions: Set<string>): NavItem[] {
  return items
    .map((item) => {
      const children = item.children ? filterNavItems(item.children, permissions) : undefined
      const allowed = !item.permission || permissions.has(item.permission)
      if (children && children.length > 0) {
        return { ...item, children }
      }
      return allowed ? { ...item, children } : null
    })
    .filter((item): item is NavItem => item !== null)
}

function toMenuItem(item: NavItem, permissions: Set<string>): MenuProps['items'][number] {
  return {
    key: item.key,
    icon: item.icon,
    label: item.label,
    children: item.children ? item.children.map((child) => toMenuItem(child, permissions)) : undefined,
  }
}
