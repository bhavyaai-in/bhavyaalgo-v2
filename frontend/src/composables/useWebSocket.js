import { ref } from 'vue'

const ws = ref(null)
let reconnectTimer = null
let listeners = {}

function getToken() {
  return localStorage.getItem('token')
}

function connect() {
  if (ws.value && (ws.value.readyState === WebSocket.OPEN || ws.value.readyState === WebSocket.CONNECTING)) return

  const token = getToken()
  if (!token) {
    reconnectTimer = setTimeout(connect, 2000)
    return
  }

  const protocol = location.protocol === 'https:' ? 'wss:' : 'ws:'
  const url = `${protocol}//${location.host}/ws?token=${token}`

  const socket = new WebSocket(url)
  ws.value = socket

  socket.onopen = () => {
    if (reconnectTimer) { clearTimeout(reconnectTimer); reconnectTimer = null }
  }

  socket.onmessage = (event) => {
    try {
      const msg = JSON.parse(event.data)
      if (msg.type === 'tick' && listeners.tick) {
        listeners.tick.forEach(fn => fn(msg.data))
      }
      if (msg.type === 'subscribed' && listeners.subscribed) {
        listeners.subscribed.forEach(fn => fn(msg.symbols))
      }
    } catch {}
  }

  socket.onclose = () => {
    if (ws.value === socket) ws.value = null
    reconnectTimer = setTimeout(connect, 3000)
  }

  socket.onerror = () => {
    socket.close()
  }
}

export function useWebSocket() {
  connect()

  function send(msg) {
    if (ws.value && ws.value.readyState === WebSocket.OPEN) {
      ws.value.send(JSON.stringify(msg))
    }
  }

  function subscribe(symbols) {
    send({ type: 'subscribe', symbols })
  }

  function unsubscribe(symbols) {
    send({ type: 'unsubscribe', symbols })
  }

  function onTick(fn) {
    if (!listeners.tick) listeners.tick = []
    listeners.tick.push(fn)
  }

  function offTick(fn) {
    if (listeners.tick) listeners.tick = listeners.tick.filter(f => f !== fn)
  }

  return { subscribe, unsubscribe, onTick, offTick, send }
}

export function disconnectWS() {
  if (ws.value) {
    ws.value.close()
    ws.value = null
  }
  if (reconnectTimer) {
    clearTimeout(reconnectTimer)
    reconnectTimer = null
  }
}
