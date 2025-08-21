<!--
/**
 * Main application template providing the global layout structure.
 * 
 * This template defines:
 * - Application header with branding and authentication controls
 * - Main content area for router-view rendering
 * - Footer with copyright and attribution
 * - Responsive design with consistent styling
 * 
 * The layout adapts based on authentication state and provides
 * a consistent user experience across all application views.
 */
-->
<template>
  <div id="app">
    <!-- Application header with branding and user controls -->
    <header class="app-header">
      <div class="header-content">
        <!-- Application title with icon -->
        <h1 class="app-title">
          <span class="icon">🎯</span>
          Intelligent Presenter
        </h1>
        
        <!-- User info and logout controls (shown when authenticated) -->
        <div class="header-actions" v-if="authStore.isAuthenticated">
          <span class="user-info">{{ authStore.userInfo?.name }}</span>
          <button @click="logout" class="logout-btn">ログアウト</button>
        </div>
      </div>
    </header>

    <!-- Main content area for view components -->
    <main class="app-main">
      <router-view />
    </main>

    <!-- Application footer with copyright -->
    <footer class="app-footer">
      <p>&copy; 2024 Intelligent Presenter - Nulab Technical Challenge</p>
    </footer>
  </div>
</template>

<!--
/**
 * @fileoverview Main application component for the Intelligent Presenter.
 * 
 * This component serves as the root of the application and provides:
 * - Global layout structure and navigation
 * - Authentication state management and initialization
 * - User session controls (login/logout)
 * - Consistent branding and styling
 * 
 * The component integrates with Pinia stores for state management and
 * Vue Router for navigation between different application views.
 * 
 * @author Nulab Technical Challenge
 * @version 1.0.0
 */
-->
<script setup lang="ts">
import { onMounted } from 'vue'
import { useAuthStore } from '@/stores/auth'
import { useRouter } from 'vue-router'

/** Authentication store for managing user sessions */
const authStore = useAuthStore()

/** Router instance for programmatic navigation */
const router = useRouter()

/**
 * Handles user logout process.
 * 
 * This function:
 * 1. Calls the auth store logout method to clear session
 * 2. Redirects user to the login page
 * 3. Ensures clean authentication state reset
 * 
 * @async
 * @function logout
 */
const logout = async () => {
  await authStore.logout()
  router.push('/login')
}

/**
 * Component initialization when mounted to DOM.
 * 
 * Initializes authentication state to restore user sessions
 * from localStorage if available. This ensures users don't
 * need to re-authenticate on page refresh.
 */
onMounted(() => {
  // Initialize authentication state from stored tokens
  authStore.initializeAuth()
})
</script>

<style scoped>
.app-header {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
  padding: 1rem 0;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
}

.header-content {
  max-width: 1200px;
  margin: 0 auto;
  padding: 0 2rem;
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.app-title {
  margin: 0;
  font-size: 1.8rem;
  font-weight: 700;
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.icon {
  font-size: 2rem;
}

.header-actions {
  display: flex;
  align-items: center;
  gap: 1rem;
}

.user-info {
  font-weight: 500;
}

.logout-btn {
  background: rgba(255, 255, 255, 0.2);
  border: 1px solid rgba(255, 255, 255, 0.3);
  color: white;
  padding: 0.5rem 1rem;
  border-radius: 6px;
  cursor: pointer;
  transition: all 0.3s ease;
}

.logout-btn:hover {
  background: rgba(255, 255, 255, 0.3);
}

.app-main {
  min-height: calc(100vh - 120px);
  padding: 2rem 0;
}

.app-footer {
  background: #f8f9fa;
  padding: 1rem 0;
  text-align: center;
  color: #666;
  border-top: 1px solid #e9ecef;
}
</style>