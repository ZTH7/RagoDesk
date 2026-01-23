import { useCallback, useEffect, useState } from 'react'

type Source = 'api' | 'fallback' | 'empty'

type RequestState<T> = {
  data: T
  loading: boolean
  error: string | null
  source: Source
  reload: () => void
}

type RequestOptions = {
  allowFallback?: boolean
  enabled?: boolean
}

export function useRequest<T>(
  fetcher: () => Promise<T>,
  fallback: T,
  options: RequestOptions = {},
): RequestState<T> {
  const [data, setData] = useState<T>(fallback)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [source, setSource] = useState<Source>('empty')
  const enabled = options.enabled !== false

  const run = useCallback(() => {
    if (!enabled) {
      setLoading(false)
      setError(null)
      setSource('empty')
      setData(fallback)
      return
    }

    let alive = true
    setLoading(true)
    setError(null)

    fetcher()
      .then((res) => {
        if (!alive) return
        setData(res)
        setSource('api')
      })
      .catch((err: Error) => {
        if (!alive) return
        setData(fallback)
        setSource(options.allowFallback ? 'fallback' : 'empty')
        setError(err.message)
      })
      .finally(() => {
        if (!alive) return
        setLoading(false)
      })

    return () => {
      alive = false
    }
  }, [enabled, fetcher, fallback, options.allowFallback])

  useEffect(() => {
    const cancel = run()
    return () => {
      if (typeof cancel === 'function') {
        cancel()
      }
    }
  }, [run])

  return {
    data,
    loading: enabled ? loading : false,
    error: enabled ? error : null,
    source,
    reload: () => {
      if (!enabled) return
      run()
    },
  }
}
