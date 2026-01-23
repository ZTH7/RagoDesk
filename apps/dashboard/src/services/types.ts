export type ApiResponse<T> = {
  code: number
  message: string
  data: T
  request_id: string
}

export type ListParams = {
  limit?: number
  offset?: number
}
