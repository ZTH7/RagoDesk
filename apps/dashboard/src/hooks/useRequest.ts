import { useCallback, useEffect, useRef, useState } from 'react'

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
  deps?: ReadonlyArray<unknown>
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
  const allowFallback = options.allowFallback ?? false
  const deps = options.deps ?? []
  const fetcherRef = useRef(fetcher)
  const fallbackRef = useRef(fallback)

  useEffect(() => {
    fetcherRef.current = fetcher
  }, [fetcher])

  useEffect(() => {
    fallbackRef.current = fallback
  }, [fallback])

  const run = useCallback(() => {
    if (!enabled) {
      setLoading(false)
      setError(null)
      setSource('empty')
      setData(fallbackRef.current)
      return
    }

    let alive = true
    setLoading(true)
    setError(null)

    fetcherRef.current()
      .then((res) => {
        if (!alive) return
        setData(res)
        setSource('api')
      })
      .catch((err: Error) => {
        if (!alive) return
        setData(fallbackRef.current)
        setSource(allowFallback ? 'fallback' : 'empty')
        setError(err.message)
      })
      .finally(() => {
        if (!alive) return
        setLoading(false)
      })

    return () => {
      alive = false
    }
  }, [allowFallback, enabled])

  useEffect(() => {
    const cancel = run()
    return () => {
      if (typeof cancel === 'function') {
        cancel()
      }
    }
  }, [run, ...deps])

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
