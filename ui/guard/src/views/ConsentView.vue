<script setup>
import { reactive, ref, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { getAppConfig } from '../utils/config.js'

const route = useRoute()
const loading = ref(false)
const error = ref('')

const params = reactive({
  response_type: '',
  client_id: '',
  redirect_uri: '',
  scope: '',
  state: '',
  nonce: '',
  code_challenge: '',
  code_challenge_method: '',
  client_name: ''
})

onMounted(() => {
  // Extract all OAuth params from query
  params.response_type = route.query.response_type || ''
  params.client_id = route.query.client_id || ''
  params.redirect_uri = route.query.redirect_uri || ''
  params.scope = route.query.scope || ''
  params.state = route.query.state || ''
  params.nonce = route.query.nonce || ''
  params.code_challenge = route.query.code_challenge || ''
  params.code_challenge_method = route.query.code_challenge_method || ''
  params.client_name = route.query.client_name || params.client_id

  // Init Matrix background
  const canvas = document.getElementById('matrix')
  if (!canvas) return
  const ctx = canvas.getContext('2d')
  canvas.width = window.innerWidth
  canvas.height = window.innerHeight

  const chars = 'ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789@#$%^&*'
  const fontSize = 14
  const columns = canvas.width / fontSize
  const drops = Array(Math.floor(columns)).fill(1)

  function draw() {
    ctx.fillStyle = 'rgba(0, 0, 0, 0.05)'
    ctx.fillRect(0, 0, canvas.width, canvas.height)
    ctx.fillStyle = '#0F0'
    ctx.font = fontSize + 'px monospace'
    for (let i = 0; i < drops.length; i++) {
      const text = chars[Math.floor(Math.random() * chars.length)]
      ctx.fillText(text, i * fontSize, drops[i] * fontSize)
      if (drops[i] * fontSize > canvas.height && Math.random() > 0.975) drops[i] = 0
      drops[i]++
    }
  }
  setInterval(draw, 33)
  window.addEventListener('resize', () => {
    canvas.width = window.innerWidth
    canvas.height = window.innerHeight
  })
})

const scopeList = () => {
  return (params.scope || 'openid').split(' ').filter(s => s)
}

const scopeDescription = (scope) => {
  const descriptions = {
    'openid': 'Verify your identity',
    'profile': 'Access your username',
    'email': 'Access your email address'
  }
  return descriptions[scope] || scope
}

const handleConsent = async (action) => {
  loading.value = true
  error.value = ''

  try {
    const config = getAppConfig()
    const baseURL = config.baseURL || 'http://localhost:9999/wall'

    // Submit consent form to OWD backend
    const formData = new URLSearchParams()
    formData.append('action', action)
    formData.append('response_type', params.response_type)
    formData.append('client_id', params.client_id)
    formData.append('redirect_uri', params.redirect_uri)
    formData.append('scope', params.scope)
    formData.append('state', params.state)
    formData.append('nonce', params.nonce)
    formData.append('code_challenge', params.code_challenge)
    formData.append('code_challenge_method', params.code_challenge_method)

    // POST to consent endpoint - follow redirects
    const response = await fetch(`${baseURL}/oauth/consent`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
      body: formData.toString(),
      credentials: 'include',
      redirect: 'follow'
    })

    if (response.redirected) {
      window.location.href = response.url
    } else if (response.ok) {
      // Check if there's a redirect in the response
      const data = await response.text()
      if (data.includes('redirect')) {
        const jsonData = JSON.parse(data)
        if (jsonData.redirect) {
          window.location.href = jsonData.redirect
        }
      }
    } else {
      error.value = 'Authorization failed. Please try again.'
    }
  } catch (e) {
    error.value = e.message || 'An error occurred'
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="consent-container">
    <canvas id="matrix"></canvas>

    <div class="consent-terminal">
      <div class="terminal-header">
        <span class="terminal-title">> AUTHORIZATION REQUEST</span>
        <div class="terminal-controls">
          <span class="control"></span>
          <span class="control"></span>
          <span class="control"></span>
        </div>
      </div>

      <div class="terminal-content">
        <div class="terminal-output">
          <div class="type-line">> Application requesting access:</div>
          <div class="app-name">{{ params.client_name }}</div>
        </div>

        <div class="scope-section">
          <div class="scope-title">> This application would like to:</div>
          <div class="scope-list">
            <div class="scope-item" v-for="scope in scopeList()" :key="scope">
              <span class="scope-check">[+]</span>
              <span class="scope-text">{{ scopeDescription(scope) }}</span>
            </div>
          </div>
        </div>

        <div class="error-text system" v-if="error">
          > ERROR: {{ error }}
        </div>

        <div class="button-row">
          <button
            class="deny-btn"
            @click="handleConsent('deny')"
            :disabled="loading"
          >
            [ DENY ]
          </button>
          <button
            class="approve-btn"
            @click="handleConsent('approve')"
            :disabled="loading"
          >
            {{ loading ? 'AUTHORIZING...' : '[ AUTHORIZE ]' }}
          </button>
        </div>
      </div>
    </div>

    <div class="copyright">
      Copyright &copy; 2023 Open Website Defender. All rights reserved.
    </div>
  </div>
</template>

<style scoped>
.consent-container {
  width: 100vw; height: 100vh; background: #000;
  display: flex; align-items: center; justify-content: center;
  font-family: 'Courier New', monospace; overflow: hidden;
}

#matrix { position: fixed; top: 0; left: 0; width: 100%; height: 100%; z-index: 1; }

.consent-terminal {
  width: 600px; background: rgba(0, 0, 0, 0.4);
  border: 1px solid rgba(0, 255, 0, 0.3);
  box-shadow: 0 0 20px rgba(0, 255, 0, 0.1), inset 0 0 30px rgba(0, 255, 0, 0.05);
  backdrop-filter: blur(3px); z-index: 2; overflow: hidden;
}

.terminal-header {
  background: rgba(0, 255, 0, 0.2); border-bottom: 1px solid rgba(0, 255, 0, 0.3);
  padding: 8px 15px; display: flex; justify-content: space-between; align-items: center;
}
.terminal-title { color: #0F0; text-shadow: 0 0 5px rgba(0, 255, 0, 0.5); }
.terminal-controls { display: flex; gap: 8px; }
.control { width: 12px; height: 12px; border-radius: 50%; background: #000; opacity: 0.5; }

.terminal-content {
  padding: 25px; color: #0F0; text-shadow: 0 0 3px rgba(0, 255, 0, 0.5);
}

.terminal-output { margin-bottom: 20px; }
.type-line { margin: 8px 0; font-size: 14px; }

.app-name {
  font-size: 22px; font-weight: bold; color: #fff;
  text-shadow: 0 0 10px rgba(255, 255, 255, 0.3);
  margin: 10px 0 0 20px;
}

.scope-section { margin: 25px 0; }
.scope-title { margin-bottom: 12px; font-size: 14px; }
.scope-list { margin-left: 20px; }
.scope-item { display: flex; align-items: center; gap: 10px; margin: 8px 0; font-size: 14px; }
.scope-check { color: #0F0; font-weight: bold; }
.scope-text { color: #ccc; }

.error-text { color: #F00; text-shadow: 0 0 5px rgba(255, 0, 0, 0.5); margin: 15px 0; font-size: 14px; }

.button-row {
  display: flex; justify-content: flex-end; gap: 15px; margin-top: 30px;
}

button {
  background: rgba(0, 255, 0, 0.1); border: 1px solid rgba(0, 255, 0, 0.3);
  color: #0F0; text-shadow: 0 0 3px rgba(0, 255, 0, 0.5);
  padding: 10px 25px; font-family: 'Courier New', monospace;
  font-size: 16px; cursor: pointer; transition: all 0.3s ease;
}
button:hover:not(:disabled) {
  background: rgba(0, 255, 0, 0.2); border-color: rgba(0, 255, 0, 0.5);
  box-shadow: 0 0 10px rgba(0, 255, 0, 0.2), inset 0 0 10px rgba(0, 255, 0, 0.2);
}
button:disabled { opacity: 0.5; cursor: not-allowed; }

.deny-btn { border-color: rgba(255, 0, 0, 0.3); color: #F44; background: rgba(255, 0, 0, 0.05); }
.deny-btn:hover:not(:disabled) {
  background: rgba(255, 0, 0, 0.15); border-color: rgba(255, 0, 0, 0.5);
  box-shadow: 0 0 10px rgba(255, 0, 0, 0.2);
}

.consent-terminal::after {
  content: ""; position: absolute; top: 0; left: 0; right: 0; height: 100%;
  background: linear-gradient(transparent 0%, rgba(0, 255, 0, 0.05) 50%, transparent 100%);
  animation: scan 8s linear infinite; pointer-events: none;
}
@keyframes scan { 0% { transform: translateY(-100%); } 100% { transform: translateY(100%); } }

.copyright {
  position: absolute; bottom: 20px; color: #0F0; font-size: 12px;
  opacity: 0.6; z-index: 10; text-shadow: 0 0 5px rgba(0, 255, 0, 0.3);
  font-family: 'Courier New', monospace;
}

@media (max-width: 650px) {
  .consent-terminal { width: 90%; margin: 20px; }
}
</style>
