<template>
  <div class="login-container">
    <canvas ref="matrixCanvas" class="matrix-bg"></canvas>

    <div class="login-glass-panel">
      <div class="terminal-header no-select">
        <span class="terminal-title">{{ t('login.system_access_required') }}</span>
        <div class="status-dots">
          <span class="dot red"></span>
          <span class="dot yellow"></span>
          <span class="dot green"></span>
        </div>
      </div>

      <div class="terminal-body">
        <!-- Step 1: Username/Password -->
        <template v-if="!authStore.requires2FA">
          <div class="system-status no-select">
            <p>{{ t('login.initializing') }}</p>
            <p>{{ t('login.establishing') }}</p>
            <p>{{ t('login.access_detected') }}</p>
            <p class="blink">{{ t('login.authenticate') }}</p>
          </div>

          <el-form
            ref="formRef"
            :model="loginForm"
            :rules="rules"
            class="login-form"
            @keyup.enter="handleLogin"
          >
            <div class="input-wrapper">
              <span class="input-label">{{ t('login.username') }}</span>
              <el-form-item prop="username" class="terminal-input">
                <el-input
                  v-model="loginForm.username"
                  autocomplete="off"
                  class="glass-input"
                />
              </el-form-item>
            </div>

            <div class="input-wrapper">
              <span class="input-label">{{ t('login.password') }}</span>
              <el-form-item prop="password" class="terminal-input">
                <el-input
                  v-model="loginForm.password"
                  type="password"
                  show-password
                  autocomplete="off"
                  class="glass-input"
                />
              </el-form-item>
            </div>

            <div class="action-area">
              <el-button
                :loading="loading"
                class="login-button"
                @click="handleLogin"
              >
                {{ t('login.btn_authenticate') }}
              </el-button>
            </div>
          </el-form>
        </template>

        <!-- Step 2: 2FA Code -->
        <template v-else>
          <div class="system-status no-select">
            <p>{{ t('login.access_detected') }}</p>
            <p class="blink">{{ t('login.two_factor_required') }}</p>
          </div>

          <div class="totp-form" @keyup.enter="handleVerify2FA">
            <div class="input-wrapper">
              <span class="input-label">{{ t('login.two_factor_code') }}</span>
              <el-input
                ref="totpInputRef"
                v-model="totpCode"
                maxlength="6"
                autocomplete="off"
                class="glass-input totp-input"
                :placeholder="'000000'"
              />
            </div>

            <div class="action-area totp-actions">
              <el-button
                class="login-button back-button"
                @click="handleCancel2FA"
              >
                {{ t('login.back') }}
              </el-button>
              <el-button
                :loading="loading"
                class="login-button"
                @click="handleVerify2FA"
              >
                {{ t('login.btn_verify') }}
              </el-button>
            </div>
          </div>
        </template>
      </div>
    </div>

    <div class="copyright">
      Copyright © 2023 舫梦科技. All rights reserved.
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted, onUnmounted, computed, nextTick, watch } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { ElMessage } from 'element-plus'
import { useI18n } from 'vue-i18n'

const router = useRouter()
const authStore = useAuthStore()
const { t } = useI18n()

const formRef = ref()
const loading = ref(false)
const matrixCanvas = ref<HTMLCanvasElement | null>(null)
const totpInputRef = ref()
const totpCode = ref('')

const loginForm = reactive({
  username: '',
  password: ''
})

const rules = computed(() => ({
  username: [{ required: true, message: t('login.required'), trigger: 'blur' }],
  password: [{ required: true, message: t('login.required'), trigger: 'blur' }]
}))

// Auto-focus TOTP input when 2FA step appears
watch(() => authStore.requires2FA, (val) => {
  if (val) {
    totpCode.value = ''
    nextTick(() => {
      totpInputRef.value?.focus()
    })
  }
})

const handleLogin = async () => {
  if (!formRef.value) return

  formRef.value.validate(async (valid: boolean) => {
    if (valid) {
      loading.value = true
      try {
        const result = await authStore.login(loginForm.username, loginForm.password)
        if (!result.requires2FA) {
          ElMessage.success(t('login.access_granted'))
          router.push({name: 'dashboard'})
        }
      } catch (error: any) {
        // Error message already shown by request interceptor
      } finally {
        loading.value = false
      }
    }
  })
}

const handleVerify2FA = async () => {
  if (totpCode.value.length !== 6) return
  loading.value = true
  try {
    await authStore.verify2FA(totpCode.value)
    ElMessage.success(t('login.access_granted'))
    router.push({name: 'dashboard'})
  } catch (error: any) {
    totpCode.value = ''
    nextTick(() => {
      totpInputRef.value?.focus()
    })
  } finally {
    loading.value = false
  }
}

const handleCancel2FA = () => {
  authStore.cancelChallenge()
  totpCode.value = ''
}

// Matrix Effect - Enhanced for Authenticity
let intervalId: any
const initMatrix = () => {
  const canvas = matrixCanvas.value
  if (!canvas) return

  const ctx = canvas.getContext('2d')
  if (!ctx) return

  canvas.width = window.innerWidth
  canvas.height = window.innerHeight

  const chars = '0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ' // Richer character set
  const letters = chars.split('')
  const fontSize = 16
  const columns = canvas.width / fontSize
  const drops: number[] = []

  for (let i = 0; i < columns; i++) {
    drops[i] = Math.floor(Math.random() * canvas.height / fontSize)
  }

  const draw = () => {
    // Very transparent black to create long trails
    ctx.fillStyle = 'rgba(0, 0, 0, 0.05)'
    ctx.fillRect(0, 0, canvas.width, canvas.height)

    ctx.fillStyle = '#0F0' // Pure Matrix Green
    ctx.font = fontSize + 'px monospace'

    for (let i = 0; i < drops.length; i++) {
      const text = letters[Math.floor(Math.random() * letters.length)]

      // Randomly brighten some characters
      if (Math.random() > 0.95) {
        ctx.fillStyle = '#FFF'
      } else {
        ctx.fillStyle = '#0F0'
      }

      ctx.fillText(text, i * fontSize, drops[i] * fontSize)

      if (drops[i] * fontSize > canvas.height && Math.random() > 0.975) {
        drops[i] = 0
      }
      drops[i]++
    }
  }

  intervalId = setInterval(draw, 33)

  window.addEventListener('resize', () => {
    canvas.width = window.innerWidth
    canvas.height = window.innerHeight
  })
}

onMounted(() => {
  initMatrix()
})

onUnmounted(() => {
  if (intervalId) clearInterval(intervalId)
})
</script>

<style scoped>
.login-container {
  height: 100vh;
  width: 100vw;
  background-color: #000;
  display: flex;
  justify-content: center;
  align-items: center;
  font-family: 'Courier New', monospace;
  overflow: hidden;
  position: relative;
}

.matrix-bg {
  position: absolute;
  top: 0;
  left: 0;
  z-index: 1;
  opacity: 0.6; /* High visibility like Guard */
}

.login-glass-panel {
  width: 550px;
  /* Strong Green Glass Effect */
  background: rgba(0, 50, 0, 0.6);
  backdrop-filter: blur(8px);
  border: 1px solid rgba(0, 255, 0, 0.3);
  box-shadow: 0 0 30px rgba(0, 255, 0, 0.2);
  position: relative;
  z-index: 10;
  padding: 30px;
  border-radius: 4px;
}

.terminal-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  border-bottom: 1px solid rgba(0, 255, 0, 0.3);
  padding-bottom: 15px;
  margin-bottom: 20px;
}

.terminal-title {
  color: #0F0;
  font-weight: bold;
  font-size: 14px;
  text-shadow: 0 0 5px rgba(0, 255, 0, 0.5);
}

.status-dots {
  display: flex;
  gap: 8px;
}

.dot {
  width: 10px;
  height: 10px;
  border-radius: 50%;
  opacity: 0.7;
}

.dot.red { background: #ff5555; }
.dot.yellow { background: #f1fa8c; }
.dot.green { background: #50fa7b; }

.system-status {
  color: #0F0;
  margin-bottom: 30px;
  line-height: 1.8;
  font-size: 14px;
  text-shadow: 0 0 3px rgba(0, 255, 0, 0.3);
}

.system-status p {
  margin: 0;
}

.blink {
  animation: blink 1s step-end infinite;
}

@keyframes blink {
  50% { opacity: 0; }
}

.input-wrapper {
  margin-bottom: 20px;
}

.input-label {
  display: block;
  color: #0F0;
  font-weight: bold;
  margin-bottom: 8px;
  font-size: 14px;
  text-shadow: 0 0 5px rgba(0, 255, 0, 0.3);
}

.terminal-input {
  margin-bottom: 0;
}

/* Custom Glass Input Styles */
:deep(.glass-input .el-input__wrapper) {
  background-color: rgba(0, 20, 0, 0.5) !important;
  box-shadow: none !important;
  border-bottom: 2px solid rgba(0, 255, 0, 0.3) !important;
  border-radius: 0;
  padding: 0 10px;
  height: 45px;
  transition: all 0.3s;
}

:deep(.glass-input .el-input__wrapper.is-focus),
:deep(.glass-input .el-input__wrapper:hover) {
  background-color: rgba(0, 40, 0, 0.6) !important;
  border-bottom-color: #0F0 !important;
  box-shadow: 0 5px 15px rgba(0, 255, 0, 0.1) !important;
}

:deep(.glass-input .el-input__inner) {
  color: #fff !important;
  font-family: 'Courier New', monospace;
  font-size: 16px;
  letter-spacing: 1px;
  height: 45px;
}

:deep(.glass-input .el-input__inner:-webkit-autofill),
:deep(.glass-input .el-input__inner:-webkit-autofill:hover),
:deep(.glass-input .el-input__inner:-webkit-autofill:focus),
:deep(.glass-input .el-input__inner:-webkit-autofill:active) {
  -webkit-text-fill-color: #fff !important;
  -webkit-background-clip: text !important;
  background-clip: text !important;
  caret-color: #fff;
}

/* TOTP login step */
:deep(.totp-input .el-input__inner) {
  font-size: 22px;
  letter-spacing: 8px;
  text-align: center;
}

.action-area {
  margin-top: 40px;
  text-align: right;
}

.totp-actions {
  display: flex;
  justify-content: space-between;
}

.login-button {
  background: transparent !important;
  border: 1px solid #0F0 !important;
  color: #0F0 !important;
  font-family: 'Courier New', monospace;
  font-weight: bold;
  font-size: 14px;
  padding: 12px 30px;
  height: auto !important;
  transition: all 0.3s;
  letter-spacing: 1px;
}

.login-button:hover {
  background: rgba(0, 255, 0, 0.2) !important;
  box-shadow: 0 0 20px rgba(0, 255, 0, 0.4);
  text-shadow: 0 0 8px #0F0;
}

.back-button {
  border-color: rgba(0, 255, 0, 0.4) !important;
  color: rgba(0, 255, 0, 0.6) !important;
}

.copyright {
  position: absolute;
  bottom: 20px;
  color: #0F0;
  font-size: 12px;
  opacity: 0.6;
  z-index: 10;
  text-shadow: 0 0 5px rgba(0, 255, 0, 0.3);
}
</style>
