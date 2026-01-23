import {
  BarChartOutlined,
  DatabaseOutlined,
  FileTextOutlined,
  HistoryOutlined,
  KeyOutlined,
  MessageOutlined,
  RobotOutlined,
  SettingOutlined,
  TeamOutlined,
  ToolOutlined,
  SafetyOutlined,
  LockOutlined,
} from '@ant-design/icons'
import { AnalyticsOverview } from '../pages/console/AnalyticsOverview'
import { AnalyticsLatency } from '../pages/console/AnalyticsLatency'
import { AnalyticsTopQuestions } from '../pages/console/AnalyticsTopQuestions'
import { AnalyticsKBGaps } from '../pages/console/AnalyticsKBGaps'
import { Users } from '../pages/console/Users'
import { Roles } from '../pages/console/Roles'
import { Permissions } from '../pages/console/Permissions'
import { Bots } from '../pages/console/Bots'
import { BotDetail } from '../pages/console/BotDetail'
import { KnowledgeBases } from '../pages/console/KnowledgeBases'
import { KnowledgeBaseDetail } from '../pages/console/KnowledgeBaseDetail'
import { Documents } from '../pages/console/Documents'
import { DocumentDetail } from '../pages/console/DocumentDetail'
import { ApiKeys } from '../pages/console/ApiKeys'
import { ApiKeyDetail } from '../pages/console/ApiKeyDetail'
import { ApiUsage } from '../pages/console/ApiUsage'
import { ApiUsageSummary } from '../pages/console/ApiUsageSummary'
import { Sessions } from '../pages/console/Sessions'
import { SessionDetail } from '../pages/console/SessionDetail'
import { DevtoolsApi } from '../pages/console/DevtoolsApi'
import { Profile } from '../pages/console/Profile'
import type { AppRoute, NavItem } from './types'
import { permissions } from '../auth/permissions'

export const consoleDefaultPath = '/console/analytics/overview'

export const consoleNavItems: NavItem[] = [
  {
    key: '/console/analytics',
    icon: <BarChartOutlined />,
    label: '统计分析',
    permission: permissions.tenant.analyticsRead,
    children: [
      { key: '/console/analytics/overview', label: '总览' },
      { key: '/console/analytics/latency', label: '延迟趋势' },
      { key: '/console/analytics/top-questions', label: 'Top Questions' },
      { key: '/console/analytics/kb-gaps', label: 'KB Gaps' },
    ],
  },
  { key: '/console/users', icon: <TeamOutlined />, label: '成员管理', permission: permissions.tenant.userRead },
  { key: '/console/roles', icon: <SafetyOutlined />, label: '角色管理', permission: permissions.tenant.roleRead },
  { key: '/console/permissions', icon: <LockOutlined />, label: '权限目录', permission: permissions.tenant.permissionRead },
  { key: '/console/bots', icon: <RobotOutlined />, label: '机器人', permission: permissions.tenant.botRead },
  { key: '/console/knowledge-bases', icon: <DatabaseOutlined />, label: '知识库', permission: permissions.tenant.knowledgeRead },
  { key: '/console/documents', icon: <FileTextOutlined />, label: '文档管理', permission: permissions.tenant.documentRead },
  { key: '/console/api-keys', icon: <KeyOutlined />, label: 'API Keys', permission: permissions.tenant.apiKeyRead },
  { key: '/console/api-usage', icon: <HistoryOutlined />, label: '调用日志', permission: permissions.tenant.apiUsageRead },
  { key: '/console/sessions', icon: <MessageOutlined />, label: '会话管理', permission: permissions.tenant.chatSessionRead },
  { key: '/console/devtools/api', icon: <ToolOutlined />, label: 'API 调试', permission: permissions.tenant.apiKeyRead },
  { key: '/console/profile', icon: <SettingOutlined />, label: '个人中心' },
]

export const consoleMenuKeys = [
  '/console/analytics/overview',
  '/console/analytics/latency',
  '/console/analytics/top-questions',
  '/console/analytics/kb-gaps',
  '/console/users',
  '/console/roles',
  '/console/permissions',
  '/console/bots',
  '/console/knowledge-bases',
  '/console/documents',
  '/console/api-keys',
  '/console/api-usage',
  '/console/sessions',
  '/console/devtools/api',
  '/console/profile',
]

export const consoleRoutes: AppRoute[] = [
  { path: 'analytics/overview', element: <AnalyticsOverview />, permission: permissions.tenant.analyticsRead },
  { path: 'analytics/latency', element: <AnalyticsLatency />, permission: permissions.tenant.analyticsRead },
  { path: 'analytics/top-questions', element: <AnalyticsTopQuestions />, permission: permissions.tenant.analyticsRead },
  { path: 'analytics/kb-gaps', element: <AnalyticsKBGaps />, permission: permissions.tenant.analyticsRead },
  { path: 'users', element: <Users />, permission: permissions.tenant.userRead },
  { path: 'roles', element: <Roles />, permission: permissions.tenant.roleRead },
  { path: 'permissions', element: <Permissions />, permission: permissions.tenant.permissionRead },
  { path: 'bots', element: <Bots />, permission: permissions.tenant.botRead },
  { path: 'bots/:id', element: <BotDetail />, permission: permissions.tenant.botRead },
  { path: 'knowledge-bases', element: <KnowledgeBases />, permission: permissions.tenant.knowledgeRead },
  { path: 'knowledge-bases/:id', element: <KnowledgeBaseDetail />, permission: permissions.tenant.knowledgeRead },
  { path: 'documents', element: <Documents />, permission: permissions.tenant.documentRead },
  { path: 'documents/:id', element: <DocumentDetail />, permission: permissions.tenant.documentRead },
  { path: 'api-keys', element: <ApiKeys />, permission: permissions.tenant.apiKeyRead },
  { path: 'api-keys/:id', element: <ApiKeyDetail />, permission: permissions.tenant.apiKeyRead },
  { path: 'api-usage', element: <ApiUsage />, permission: permissions.tenant.apiUsageRead },
  { path: 'api-usage/summary', element: <ApiUsageSummary />, permission: permissions.tenant.apiUsageRead },
  { path: 'sessions', element: <Sessions />, permission: permissions.tenant.chatSessionRead },
  { path: 'sessions/:id', element: <SessionDetail />, permission: permissions.tenant.chatSessionRead },
  { path: 'devtools/api', element: <DevtoolsApi />, permission: permissions.tenant.apiKeyRead },
  { path: 'profile', element: <Profile /> },
]
