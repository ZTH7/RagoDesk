export const permissions = {
  tenant: {
    analyticsRead: 'tenant.analytics.read',
    userRead: 'tenant.user.read',
    roleRead: 'tenant.role.read',
    permissionRead: 'tenant.permission.read',
    botRead: 'tenant.bot.read',
    knowledgeRead: 'tenant.knowledge_base.read',
    documentRead: 'tenant.document.read',
    apiKeyRead: 'tenant.api_key.read',
    apiUsageRead: 'tenant.api_usage.read',
    chatSessionRead: 'tenant.chat_session.read',
  },
  platform: {
    tenantRead: 'platform.tenant.read',
    adminRead: 'platform.admin.read',
    roleRead: 'platform.role.read',
    permissionRead: 'platform.permission.read',
  },
}

export const defaultConsolePermissions = new Set<string>([
  permissions.tenant.analyticsRead,
  permissions.tenant.userRead,
  permissions.tenant.roleRead,
  permissions.tenant.permissionRead,
  permissions.tenant.botRead,
  permissions.tenant.knowledgeRead,
  permissions.tenant.documentRead,
  permissions.tenant.apiKeyRead,
  permissions.tenant.apiUsageRead,
  permissions.tenant.chatSessionRead,
])

export const defaultPlatformPermissions = new Set<string>([
  permissions.platform.tenantRead,
  permissions.platform.adminRead,
  permissions.platform.roleRead,
  permissions.platform.permissionRead,
])
