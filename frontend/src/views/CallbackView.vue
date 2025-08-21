<template>
  <div class="callback-view">
    <div class="callback-container">
      <div class="callback-card">
        <div v-if="isProcessing" class="processing">
          <div class="spinner-large"></div>
          <h2>認証処理中...</h2>
          <p>Backlogからの認証情報を処理しています。</p>
        </div>

        <div v-else-if="error" class="error">
          <div class="error-icon">❌</div>
          <h2>認証エラー</h2>
          <p>{{ error }}</p>
          <button @click="goToLogin" class="retry-button">
            ログインページに戻る
          </button>
        </div>

        <div v-else class="success">
          <div class="success-icon">✅</div>
          <h2>認証成功</h2>
          <p>Intelligent Presenterにリダイレクトしています...</p>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

const router = useRouter()
const route = useRoute()
const authStore = useAuthStore()

const isProcessing = ref(true)
const error = ref<string | null>(null)

const goToLogin = () => {
  router.push('/login')
}

onMounted(async () => {
  try {
    const token = route.query.token as string
    const success = route.query.success as string
    const code = route.query.code as string
    const state = route.query.state as string

    console.log('Callback parameters:', { token: !!token, success, code: !!code, state: !!state })

    if (token && success === 'true') {
      // 直接从 URL 参数获取 token（后端重定向方式）
      console.log('Processing token from URL parameters')
      authStore.setToken(token)
      
      // 获取用户信息
      await authStore.initializeAuth()
      console.log('Auth initialization completed')
    } else if (code && state) {
      // 传统的回调方式（如果需要的话）
      console.log('Processing traditional OAuth callback')
      await authStore.handleCallback(code, state)
    } else {
      console.error('Missing authentication parameters:', { token, success, code, state })
      throw new Error('認証コードが見つかりません')
    }
    
    // Success - redirect to home page after a short delay
    setTimeout(() => {
      router.push('/')
    }, 1500)
    
  } catch (err) {
    console.error('Callback error:', err)
    error.value = err instanceof Error ? err.message : '認証に失敗しました'
  } finally {
    isProcessing.value = false
  }
})
</script>

<style scoped>
.callback-view {
  min-height: 100vh;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 2rem;
}

.callback-container {
  width: 100%;
  max-width: 400px;
}

.callback-card {
  background: white;
  border-radius: 16px;
  box-shadow: 0 20px 40px rgba(0, 0, 0, 0.1);
  padding: 3rem 2rem;
  text-align: center;
}

.processing, .error, .success {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 1rem;
}

.spinner-large {
  width: 48px;
  height: 48px;
  border: 4px solid #f3f3f3;
  border-top: 4px solid #667eea;
  border-radius: 50%;
  animation: spin 1s linear infinite;
}

@keyframes spin {
  0% { transform: rotate(0deg); }
  100% { transform: rotate(360deg); }
}

.error-icon, .success-icon {
  font-size: 3rem;
}

h2 {
  margin: 0;
  color: #333;
  font-size: 1.5rem;
}

p {
  margin: 0;
  color: #666;
  line-height: 1.5;
}

.retry-button {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
  border: none;
  padding: 0.75rem 1.5rem;
  border-radius: 8px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.3s ease;
  margin-top: 1rem;
}

.retry-button:hover {
  transform: translateY(-2px);
  box-shadow: 0 8px 20px rgba(102, 126, 234, 0.3);
}
</style>