<script setup>
import {nextTick, onMounted, reactive, ref} from 'vue'
import {useRoute, useRouter} from 'vue-router'
import request from '@/utils/request'
import {getAppConfig} from "../utils/config.js";

const router = useRouter()
const route = useRoute()
const loading = ref(false)
const requires2FA = ref(false)
const challengeToken = ref('')
const totpCode = ref('')
const totpInputRef = ref(null)

const formData = reactive({
  username: '',
  password: '',
  rememberMe: false
})

const errors = reactive({
  username: '',
  password: '',
  general: ''
})

const validateForm = () => {
  let isValid = true
  errors.username = ''
  errors.password = ''
  errors.general = ''

  if (!formData.username) {
    errors.username = 'username is required'
    isValid = false
  }

  if (!formData.password) {
    errors.password = 'password is required'
    isValid = false
  } else if (formData.password.length < 6) {
    errors.password = 'password must be at least 6 characters long'
    isValid = false
  }

  return isValid
}


// 检查重定向参数
const checkRedirectParam = () => {
  if (!route.query.redirect) {
    errors.general = '无效的访问，请从正确的入口访问'
    return false
  }
  return true
}

// 修改组件挂载时的检查逻辑
onMounted(async () => {
  // if (!checkRedirectParam()) {
  //   return
  // }
  // await checkExistingToken()

  // 添加 Matrix 背景效果
  const canvas = document.getElementById('matrix')
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

      if (drops[i] * fontSize > canvas.height && Math.random() > 0.975) {
        drops[i] = 0
      }
      drops[i]++
    }
  }

  setInterval(draw, 33)

  window.addEventListener('resize', () => {
    canvas.width = window.innerWidth
    canvas.height = window.innerHeight
  })
})

const completeLogin = (token) => {
  localStorage.setItem('flmelody.token', token)

  const expires = new Date(Date.now() + 24 * 60 * 60 * 1000)
  const guardDomain = getAppConfig().guardDomain
  const domainPart = guardDomain && guardDomain.includes('.') ? `;domain=${guardDomain}` : ''
  document.cookie = `flmelody.token=${token}; expires=${expires.toUTCString()}; path=/${domainPart}`

  if (formData.rememberMe) {
    localStorage.setItem('rememberMe', 'true')
  }

  request.defaults.headers.common['Defender-Authorization'] = `Bearer ${token}`

  if (!route.query.redirect) {
    errors.general = '无效的访问，请从正确的入口访问'
    return
  }
  const redirectUrl = route.query.redirect
  window.location.href = decodeURIComponent(redirectUrl)
}

const handleLogin = async () => {
  if (!validateForm()) return

  loading.value = true
  errors.general = ''

  try {
    const data = await request.post('/login', {
      username: formData.username,
      password: formData.password
    })

    if (data.requires_two_fa) {
      challengeToken.value = data.challenge_token
      requires2FA.value = true
      nextTick(() => {
        totpInputRef.value?.focus()
      })
    } else if (data.token) {
      completeLogin(data.token)
    }
  } catch (error) {
    errors.general = error.message || '登录失败，请重试'
  } finally {
    loading.value = false
  }
}

const handleVerify2FA = async () => {
  if (totpCode.value.length !== 6) return

  loading.value = true
  errors.general = ''

  try {
    const data = await request.post('/login/2fa', {
      challenge_token: challengeToken.value,
      code: totpCode.value
    })

    if (data.token) {
      completeLogin(data.token)
    }
  } catch (error) {
    errors.general = error.message || 'Verification failed'
    totpCode.value = ''
    nextTick(() => {
      totpInputRef.value?.focus()
    })
  } finally {
    loading.value = false
  }
}

const handleCancel2FA = () => {
  requires2FA.value = false
  challengeToken.value = ''
  totpCode.value = ''
  errors.general = ''
}
</script>

<template>
  <div class="login-container">
    <!-- Matrix背景 -->
    <canvas id="matrix" ref="matrix"></canvas>

    <div class="login-terminal">
      <div class="terminal-header">
        <span class="terminal-title">> SYSTEM ACCESS REQUIRED</span>
        <div class="terminal-controls">
          <span class="control"></span>
          <span class="control"></span>
          <span class="control"></span>
        </div>
      </div>

      <div class="terminal-content">
        <!-- Step 1: Username/Password -->
        <template v-if="!requires2FA">
          <div class="terminal-output">
            <div class="type-line">> Initializing security protocol...</div>
            <div class="type-line">> Establishing secure connection...</div>
            <div class="type-line">> Access point detected...</div>
            <div class="type-line">> Please authenticate to continue:</div>
          </div>

          <form @submit.prevent="handleLogin" class="login-form">
            <div class="input-line">
              <span class="prompt">> Username:</span>
              <input
                type="text"
                v-model="formData.username"
                :class="{ 'error': errors.username }"
                autocomplete="off"
                spellcheck="false"
              >
            </div>
            <div class="error-text" v-if="errors.username">{{ errors.username }}</div>

            <div class="input-line">
              <span class="prompt">> Password:</span>
              <input
                type="password"
                v-model="formData.password"
                :class="{ 'error': errors.password }"
              >
            </div>
            <div class="error-text" v-if="errors.password">{{ errors.password }}</div>

            <div class="error-text system" v-if="errors.general">
              > ERROR: {{ errors.general }}
            </div>

            <div class="input-line submit-line">
              <button
                type="submit"
                :disabled="loading"
              >
                {{ loading ? 'AUTHENTICATING...' : '[ AUTHENTICATE ]' }}
              </button>
            </div>
          </form>
        </template>

        <!-- Step 2: 2FA Code -->
        <template v-else>
          <div class="terminal-output">
            <div class="type-line">> Access point detected...</div>
            <div class="type-line">> Two-factor authentication required:</div>
          </div>

          <form @submit.prevent="handleVerify2FA" class="login-form">
            <div class="input-line">
              <span class="prompt">> 2FA Code:</span>
              <input
                ref="totpInputRef"
                type="text"
                v-model="totpCode"
                maxlength="6"
                autocomplete="off"
                spellcheck="false"
                class="totp-input"
                placeholder="000000"
              >
            </div>

            <div class="error-text system" v-if="errors.general">
              > ERROR: {{ errors.general }}
            </div>

            <div class="input-line submit-line totp-actions">
              <button
                type="button"
                @click="handleCancel2FA"
                class="back-button"
              >
                [ BACK ]
              </button>
              <button
                type="submit"
                :disabled="loading || totpCode.length !== 6"
              >
                {{ loading ? 'VERIFYING...' : '[ VERIFY ]' }}
              </button>
            </div>
          </form>
        </template>
      </div>
    </div>

    <div class="copyright">
      Copyright © 2023 舫梦科技. All rights reserved.
    </div>
  </div>
</template>

<style scoped>
.login-container {
  width: 100vw;
  height: 100vh;
  background: #000;
  display: flex;
  align-items: center;
  justify-content: center;
  font-family: 'Courier New', monospace;
  overflow: hidden;
}

#matrix {
  position: fixed;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  z-index: 1;
}

.login-terminal {
  width: 800px;
  background: rgba(0, 0, 0, 0.4);
  border: 1px solid rgba(0, 255, 0, 0.3);
  box-shadow: 0 0 20px rgba(0, 255, 0, 0.1),
  inset 0 0 30px rgba(0, 255, 0, 0.05);
  backdrop-filter: blur(3px);
  z-index: 2;
  overflow: hidden;
}

.terminal-header {
  background: rgba(0, 255, 0, 0.2);
  border-bottom: 1px solid rgba(0, 255, 0, 0.3);
  padding: 8px 15px;
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.terminal-title {
  color: #0F0;
  text-shadow: 0 0 5px rgba(0, 255, 0, 0.5);
}

.terminal-controls {
  display: flex;
  gap: 8px;
}

.control {
  width: 12px;
  height: 12px;
  border-radius: 50%;
  background: #000;
  opacity: 0.5;
}

.terminal-content {
  padding: 20px;
  color: #0F0;
  text-shadow: 0 0 3px rgba(0, 255, 0, 0.5);
}

.terminal-output {
  margin-bottom: 30px;
}

.type-line {
  margin: 8px 0;
  animation: typing 0.5s steps(40, end);
}

@keyframes typing {
  from {
    width: 0
  }
  to {
    width: 100%
  }
}

.input-line {
  display: flex;
  align-items: center;
  margin: 15px 0;
  gap: 10px;
}

.prompt {
  color: #0F0;
  white-space: nowrap;
}

input {
  background: rgba(0, 0, 0, 0.3);
  border: none;
  border-bottom: 1px solid rgba(0, 255, 0, 0.3);
  color: #0F0;
  text-shadow: 0 0 3px rgba(0, 255, 0, 0.5);
  font-family: 'Courier New', monospace;
  font-size: 16px;
  padding: 5px 10px;
  flex-grow: 1;
  transition: all 0.3s ease;
}

input:focus {
  outline: none;
  border-bottom-color: rgba(0, 255, 0, 0.8);
  background: rgba(0, 255, 0, 0.1);
}

input.error {
  border-bottom-color: #F00;
  color: #F00;
}

input:-webkit-autofill,
input:-webkit-autofill:hover,
input:-webkit-autofill:focus,
input:-webkit-autofill:active {
  -webkit-text-fill-color: #0F0 !important;
  -webkit-background-clip: text !important;
  background-clip: text !important;
  caret-color: #0F0;
}

.error-text {
  color: #F00;
  text-shadow: 0 0 5px rgba(255, 0, 0, 0.5);
  opacity: 0.8;
  margin-left: 120px;
  font-size: 14px;
  animation: blink 1s infinite;
}

.error-text.system {
  margin-left: 0;
  margin-top: 15px;
  margin-bottom: 15px;
}

@keyframes blink {
  50% {
    opacity: 0.5;
  }
}

.checkbox {
  display: flex;
  align-items: center;
  gap: 10px;
  cursor: pointer;
}

.checkbox input {
  display: none;
}

.checkbox-text {
  color: #0F0;
}

.checkbox input:checked + .checkbox-text {
  color: #0F0;
  text-shadow: 0 0 5px #0F0;
}

button {
  background: rgba(0, 255, 0, 0.1);
  border: 1px solid rgba(0, 255, 0, 0.3);
  color: #0F0;
  text-shadow: 0 0 3px rgba(0, 255, 0, 0.5);
  padding: 10px 20px;
  font-family: 'Courier New', monospace;
  font-size: 16px;
  cursor: pointer;
  transition: all 0.3s ease;
}

button:hover:not(:disabled) {
  background: rgba(0, 255, 0, 0.2);
  border-color: rgba(0, 255, 0, 0.5);
  box-shadow: 0 0 10px rgba(0, 255, 0, 0.2),
  inset 0 0 10px rgba(0, 255, 0, 0.2);
}

button:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.submit-line {
  justify-content: flex-end;
  margin-top: 30px;
}

.totp-actions {
  justify-content: space-between;
}

.totp-input {
  font-size: 22px !important;
  letter-spacing: 8px;
  text-align: center;
}

.back-button {
  border-color: rgba(0, 255, 0, 0.2);
  color: rgba(0, 255, 0, 0.6);
}

@media (max-width: 850px) {
  .login-terminal {
    width: 90%;
    margin: 20px;
  }

  .input-line {
    flex-direction: column;
    align-items: flex-start;
  }

  .error-text {
    margin-left: 0;
  }
}

/* 添加扫描线动画效果 */
.login-terminal::after {
  content: "";
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  height: 100%;
  background: linear-gradient(
    transparent 0%,
    rgba(0, 255, 0, 0.05) 50%,
    transparent 100%
  );
  animation: scan 8s linear infinite;
  pointer-events: none;
}

@keyframes scan {
  0% {
    transform: translateY(-100%);
  }
  100% {
    transform: translateY(100%);
  }
}

.copyright {
  position: absolute;
  bottom: 20px;
  color: #0F0;
  font-size: 12px;
  opacity: 0.6;
  z-index: 10;
  text-shadow: 0 0 5px rgba(0, 255, 0, 0.3);
  font-family: 'Courier New', monospace;
}
</style>
