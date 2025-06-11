// WebSocket 连接管理
class WebSocketClient {
    constructor() {
        this.ws = null
        this.url = ''
        this.reconnectInterval = 5000
        this.maxReconnectAttempts = 5
        this.reconnectAttempts = 0
        this.listeners = new Map()
    }

    connect(url) {
        this.url = url

        try {
            this.ws = new WebSocket(url)

            this.ws.onopen = () => {
                console.log('WebSocket connected')
                this.reconnectAttempts = 0
                this.emit('connected')
            }

            this.ws.onmessage = (event) => {
                try {
                    const data = JSON.parse(event.data)
                    this.emit('message', data)

                    // 分发特定类型的消息
                    if (data.type) {
                        this.emit(data.type, data.payload)
                    }
                } catch (error) {
                    console.error('Failed to parse WebSocket message:', error)
                }
            }

            this.ws.onclose = () => {
                console.log('WebSocket disconnected')
                this.emit('disconnected')
                this.reconnect()
            }

            this.ws.onerror = (error) => {
                console.error('WebSocket error:', error)
                this.emit('error', error)
            }
        } catch (error) {
            console.error('Failed to create WebSocket connection:', error)
            this.reconnect()
        }
    }

    reconnect() {
        if (this.reconnectAttempts < this.maxReconnectAttempts) {
            this.reconnectAttempts++
            console.log(`Attempting to reconnect... (${this.reconnectAttempts}/${this.maxReconnectAttempts})`)

            setTimeout(() => {
                this.connect(this.url)
            }, this.reconnectInterval)
        } else {
            console.error('Max reconnection attempts reached')
            this.emit('maxReconnectReached')
        }
    }

    send(data) {
        if (this.ws && this.ws.readyState === WebSocket.OPEN) {
            this.ws.send(JSON.stringify(data))
        } else {
            console.warn('WebSocket is not connected')
        }
    }

    on(event, callback) {
        if (!this.listeners.has(event)) {
            this.listeners.set(event, [])
        }
        this.listeners.get(event).push(callback)
    }

    off(event, callback) {
        if (this.listeners.has(event)) {
            const callbacks = this.listeners.get(event)
            const index = callbacks.indexOf(callback)
            if (index > -1) {
                callbacks.splice(index, 1)
            }
        }
    }

    emit(event, data) {
        if (this.listeners.has(event)) {
            this.listeners.get(event).forEach(callback => {
                try {
                    callback(data)
                } catch (error) {
                    console.error(`Error in event listener for ${event}:`, error)
                }
            })
        }
    }

    disconnect() {
        if (this.ws) {
            this.ws.close()
            this.ws = null
        }
        this.listeners.clear()
    }
}

// 单例模式
const wsClient = new WebSocketClient()

export default wsClient
