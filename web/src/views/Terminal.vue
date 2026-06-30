<template>
  <div class="terminal">
    <el-card>
      <template #header>
        <span>终端</span>
        <el-button @click="connect" text style="float: right">{{ connected ? '断开' : '连接' }}</el-button>
      </template>
      <div ref="terminalContainer" class="terminal-container"></div>
    </el-card>
  </div>
</template>

<script setup>
import { ref, onMounted, onUnmounted, nextTick } from 'vue'
import { Terminal } from '@xterm/xterm'
import { FitAddon } from '@xterm/addon-fit'
import { WebLinksAddon } from '@xterm/addon-web-links'
import '@xterm/xterm/css/xterm.css'

const terminalContainer = ref(null)
const connected = ref(false)
let terminal = null
let fitAddon = null
let ws = null

onMounted(async () => {
  await nextTick()
  initTerminal()
})

onUnmounted(() => {
  disconnect()
  if (terminal) {
    terminal.dispose()
  }
})

const initTerminal = () => {
  terminal = new Terminal({
    fontSize: 14,
    fontFamily: 'monospace',
    theme: {
      background: '#1e1e1e',
      foreground: '#d4d4d4',
      cursor: '#d4d4d4',
      selectionBackground: '#264f78'
    }
  })

  fitAddon = new FitAddon()
  terminal.loadAddon(fitAddon)
  terminal.loadAddon(new WebLinksAddon())

  terminal.open(terminalContainer.value)
  fitAddon.fit()

  window.addEventListener('resize', () => {
    fitAddon.fit()
  })
}

const connect = () => {
  if (connected.value) {
    disconnect()
    return
  }

  const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
  const wsUrl = `${protocol}//${window.location.host}/ws/terminal`
  console.log('Connecting to:', wsUrl)
  ws = new WebSocket(wsUrl)

  ws.onopen = () => {
    connected.value = true
    terminal.write('Connected to terminal\r\n')
  }

  ws.onerror = (error) => {
    console.error('WebSocket error:', error)
    terminal.write('\r\nConnection error\r\n')
  }

  ws.onmessage = (event) => {
    if (event.data instanceof Blob) {
      event.data.arrayBuffer().then(buffer => {
        const text = new TextDecoder().decode(buffer)
        terminal.write(text)
      })
    } else {
      terminal.write(event.data)
    }
  }

  ws.onclose = () => {
    connected.value = false
    terminal.write('\r\nConnection closed\r\n')
  }

  terminal.onData((data) => {
    if (ws && ws.readyState === WebSocket.OPEN) {
      ws.send(data)
    }
  })
}

const disconnect = () => {
  if (ws) {
    ws.close()
    ws = null
  }
  connected.value = false
}
</script>

<style scoped>
.terminal {
  padding: 20px;
  height: calc(100vh - 180px);
}

.terminal-container {
  height: calc(100vh - 260px);
}
</style>
