export function normalizeAccount(raw: string) {
  return raw.trim()
}

export function normalizePhone(raw: string) {
  return raw.replace(/[\s-]/g, '')
}

export function looksLikeEmail(value: string) {
  return /\S+@\S+\.\S+/.test(value)
}

export function looksLikePhone(value: string) {
  return /^\+?\d{6,20}$/.test(normalizePhone(value))
}

export function validateAccount(value: string) {
  const normalized = normalizeAccount(value)
  return looksLikeEmail(normalized) || looksLikePhone(normalized)
}
