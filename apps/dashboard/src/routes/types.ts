import type { ReactNode } from 'react'

export type AppRoute = {
  path: string
  element: ReactNode
  permission?: string
}

export type NavItem = {
  key: string
  label: ReactNode
  icon?: ReactNode
  permission?: string
  children?: NavItem[]
}
