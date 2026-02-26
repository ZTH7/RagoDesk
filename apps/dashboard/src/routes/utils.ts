import type { MenuProps } from 'antd'
import type { ItemType } from 'antd/es/menu/interface'
import type { NavItem } from './types'

export function resolveSelectedKey(pathname: string, menuKeys: string[]) {
  const match = menuKeys
    .filter((key) => pathname.startsWith(key))
    .sort((a, b) => b.length - a.length)[0]
  return match ?? menuKeys[0]
}

export function buildMenuItems(items: NavItem[], permissions: Set<string>): ItemType[] {
  const filtered = filterNavItems(items, permissions)
  return filtered.map((item) => toMenuItem(item))
}

function filterNavItems(items: NavItem[], permissions: Set<string>): NavItem[] {
  const result: NavItem[] = []
  for (const item of items) {
    const children = item.children ? filterNavItems(item.children, permissions) : undefined
    const allowed = !item.permission || permissions.has(item.permission)

    if (children && children.length > 0) {
      result.push({ ...item, children })
      continue
    }

    if (allowed) {
      result.push({ ...item, children: undefined })
    }
  }
  return result
}

function toMenuItem(item: NavItem): ItemType {
  const menuItem: NonNullable<MenuProps['items']>[number] = {
    key: item.key,
    icon: item.icon,
    label: item.label,
    children: item.children ? item.children.map((child) => toMenuItem(child)) : undefined,
  }
  return menuItem
}
