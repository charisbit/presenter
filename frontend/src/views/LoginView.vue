<template>
  <div class="login-view">
    <div class="login-container">
      <div class="login-card">
        <div class="login-header">
          <h1 class="login-title">
            <span class="icon">🎯</span>
            Intelligent Presenter
          </h1>
          <p class="login-subtitle">Backlogプロジェクトから自動でプレゼンテーションを生成</p>
        </div>

        <div class="login-content">
          <div class="login-description">
            <h2>主な機能</h2>
            <ul class="feature-list">
              <li>📊 10種類のスライドテーマ自動生成</li>
              <li>🎤 日本語音声解説付き</li>
              <li>📈 Chart.js + Mermaid図表統合</li>
              <li>⚡ リアルタイム流式生成</li>
              <li>🔐 OAuth 2.0セキュア認証</li>
            </ul>
          </div>

          <div class="login-actions">
            <button 
              @click="handleLogin" 
              :disabled="authStore.isLoading"
              class="login-button"
            >
              <span v-if="!authStore.isLoading">
                <span class="button-icon">🔑</span>
                Backlogでログイン
              </span>
              <span v-else class="loading">
                <span class="spinner"></span>
                認証中...
              </span>
            </button>
          </div>

          <div class="login-info">
            <p class="info-text">
              ⚠️ このアプリケーションを使用するには、Backlogアカウントが必要です。
              OAuth 2.0認証により、安全にログインできます。
            </p>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useAuthStore } from '@/stores/auth'

const authStore = useAuthStore()

const handleLogin = async () => {
  try {
    const authUrl = await authStore.login()
    // Redirect to Backlog OAuth page
    window.location.href = authUrl
  } catch (error) {
    console.error('Login failed:', error)
    alert('ログインに失敗しました。もう一度お試しください。')
  }
}
</script>

<style scoped>
.login-view {
  min-height: 100vh;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 2rem;
}

.login-container {
  width: 100%;
  max-width: 480px;
}

.login-card {
  background: white;
  border-radius: 16px;
  box-shadow: 0 20px 40px rgba(0, 0, 0, 0.1);
  overflow: hidden;
}

.login-header {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
  padding: 2rem;
  text-align: center;
}

.login-title {
  margin: 0 0 0.5rem 0;
  font-size: 2rem;
  font-weight: 700;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 0.5rem;
}

.icon {
  font-size: 2.5rem;
}

.login-subtitle {
  margin: 0;
  opacity: 0.9;
  font-size: 1rem;
}

.login-content {
  padding: 2rem;
}

.login-description h2 {
  margin: 0 0 1rem 0;
  color: #333;
  font-size: 1.2rem;
}

.feature-list {
  list-style: none;
  padding: 0;
  margin: 0 0 2rem 0;
}

.feature-list li {
  padding: 0.5rem 0;
  color: #666;
  font-size: 0.95rem;
}

.login-actions {
  margin: 2rem 0;
}

.login-button {
  width: 100%;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
  border: none;
  padding: 1rem 2rem;
  border-radius: 8px;
  font-size: 1.1rem;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.3s ease;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 0.5rem;
}

.login-button:hover:not(:disabled) {
  transform: translateY(-2px);
  box-shadow: 0 8px 20px rgba(102, 126, 234, 0.3);
}

.login-button:disabled {
  opacity: 0.7;
  cursor: not-allowed;
  transform: none;
}

.button-icon {
  font-size: 1.2rem;
}

.loading {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.spinner {
  width: 16px;
  height: 16px;
  border: 2px solid rgba(255, 255, 255, 0.3);
  border-top: 2px solid white;
  border-radius: 50%;
  animation: spin 1s linear infinite;
}

@keyframes spin {
  0% { transform: rotate(0deg); }
  100% { transform: rotate(360deg); }
}

.login-info {
  background: #f8f9fa;
  border-radius: 8px;
  padding: 1rem;
  margin-top: 1.5rem;
}

.info-text {
  margin: 0;
  font-size: 0.9rem;
  color: #666;
  line-height: 1.5;
}
</style>