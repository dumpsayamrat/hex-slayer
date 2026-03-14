import { useState, useEffect, useRef, useCallback } from 'react'

const WS_URL = `${window.location.protocol === 'https:' ? 'wss:' : 'ws:'}//${window.location.host}/ws`

export default function useGameSocket(token) {
  const [connected, setConnected] = useState(false)
  const [lastMessage, setLastMessage] = useState(null)
  const wsRef = useRef(null)
  const reconnectTimer = useRef(null)

  const connect = useCallback(() => {
    if (!token) return

    const ws = new WebSocket(`${WS_URL}?token=${token}`)

    ws.onopen = () => {
      console.log('[WS] connected')
      setConnected(true)
    }

    ws.onmessage = (event) => {
      const data = JSON.parse(event.data)
      console.log('[WS] message:', data)
      setLastMessage(data)
    }

    ws.onclose = () => {
      console.log('[WS] disconnected, reconnecting in 3s...')
      setConnected(false)
      reconnectTimer.current = setTimeout(connect, 3000)
    }

    ws.onerror = (err) => {
      console.error('[WS] error:', err)
      ws.close()
    }

    wsRef.current = ws
  }, [token])

  useEffect(() => {
    connect()
    return () => {
      clearTimeout(reconnectTimer.current)
      wsRef.current?.close()
    }
  }, [connect])

  const sendMessage = useCallback((msg) => {
    if (wsRef.current?.readyState === WebSocket.OPEN) {
      wsRef.current.send(JSON.stringify(msg))
    }
  }, [])

  return { connected, lastMessage, sendMessage }
}
