/**
 * @fileoverview Vue Router configuration for the Intelligent Presenter application.
 * 
 * This module sets up client-side routing with authentication guards and defines
 * all application routes. It includes protection for authenticated and guest-only
 * routes, automatic redirects, and proper route parameter handling.
 * 
 * @author Technical Challenge
 * @version 1.0.0
 */

import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import ProjectSelectionView from '@/views/ProjectSelectionView.vue'
import LoginView from '@/views/LoginView.vue'
import PresentationView from '@/views/PresentationView.vue'
import CallbackView from '@/views/CallbackView.vue'

/**
 * Create and configure the Vue Router instance.
 * Uses HTML5 history mode for clean URLs without hash fragments.
 * 
 * Route Structure:
 * - `/` - Project selection (authenticated users only)
 * - `/login` - Authentication page (guest users only)  
 * - `/auth/callback` - OAuth callback handler (public)
 * - `/presentation/:slideId` - Slide presentation viewer (authenticated users only)
 * - `*` - Catch-all redirect to home page
 */
const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: '/',
      name: 'project-selection',
      component: ProjectSelectionView,
      meta: { 
        requiresAuth: true,
        title: 'Project Selection',
        description: 'Select a Backlog project to generate presentation slides'
      }
    },
    {
      path: '/login',
      name: 'login',
      component: LoginView,
      meta: { 
        requiresGuest: true,
        title: 'Login',
        description: 'Authenticate with Backlog to access the application'
      }
    },
    {
      path: '/auth/callback',
      name: 'callback',
      component: CallbackView,
      meta: {
        title: 'Authentication Callback',
        description: 'Processing authentication response from Backlog'
      }
    },
    {
      path: '/presentation/:slideId',
      name: 'presentation',
      component: PresentationView,
      meta: { 
        requiresAuth: true,
        title: 'Presentation',
        description: 'View and present generated slides'
      },
      props: true
    },
    {
      path: '/:pathMatch(.*)*',
      redirect: '/'
    }
  ]
})

/**
 * Global navigation guard for authentication and authorization.
 * 
 * This guard runs before every route navigation and:
 * 1. Checks if the target route requires authentication
 * 2. Redirects unauthenticated users to login page
 * 3. Redirects authenticated users away from guest-only pages
 * 4. Allows navigation for authorized users
 * 
 * @param to - Target route being navigated to
 * @param from - Current route being navigated from  
 * @param next - Function to continue or redirect navigation
 * 
 * @example
 * ```typescript
 * // Authenticated user trying to access login page
 * // Will be redirected to project selection
 * 
 * // Unauthenticated user trying to access presentation
 * // Will be redirected to login page
 * ```
 */
router.beforeEach((to, from, next) => {
  const authStore = useAuthStore()
  
  // Check if route requires authentication
  if (to.meta.requiresAuth && !authStore.isAuthenticated) {
    // Redirect to login page if not authenticated
    next('/login')
  } else if (to.meta.requiresGuest && authStore.isAuthenticated) {
    // Redirect to home page if already authenticated
    next('/')
  } else {
    // Allow navigation
    next()
  }
})

/**
 * Export the configured router instance.
 * This router will be installed in the main Vue application.
 */
export default router