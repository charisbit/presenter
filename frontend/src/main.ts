/**
 * @fileoverview Main application entry point for the Intelligent Presenter frontend.
 * 
 * This module initializes the Vue 3 application with all necessary plugins and dependencies:
 * - Pinia for state management
 * - Vue Router for navigation
 * - Chart.js for data visualization components
 * - Global styles and configurations
 * 
 * The application provides an intelligent presentation generation system that integrates
 * with Backlog projects to create automated slide presentations with AI-powered content
 * generation and text-to-speech capabilities.
 * 
 * @author Technical Challenge
 * @version 1.0.0
 */

import { createApp } from 'vue'
import { createPinia } from 'pinia'
import App from './App.vue'
import router from './router'

// Import global CSS styles
import './style.css'

// Import markdown logger to ensure it's included in build
import './utils/markdownLogger'

// Import Chart.js components for data visualization
import {
  Chart as ChartJS,
  CategoryScale,
  LinearScale,
  PointElement,
  LineElement,
  BarElement,
  Title,
  Tooltip,
  Legend,
  ArcElement
} from 'chart.js'

/**
 * Register Chart.js components globally for use in slide visualizations.
 * These components enable rendering of various chart types including:
 * - Line charts for trend analysis
 * - Bar charts for comparisons
 * - Pie charts for distributions
 * - Mixed chart types for complex data presentation
 */
ChartJS.register(
  CategoryScale,
  LinearScale,
  PointElement,
  LineElement,
  BarElement,
  Title,
  Tooltip,
  Legend,
  ArcElement
)

/**
 * Create the main Vue application instance.
 * This serves as the root container for the entire frontend application.
 */
const app = createApp(App)

/**
 * Install Pinia state management plugin.
 * Pinia provides reactive state management for:
 * - Authentication state
 * - Slide generation and management
 * - Real-time WebSocket communication
 * - User preferences and settings
 */
app.use(createPinia())

/**
 * Install Vue Router for client-side navigation.
 * The router handles navigation between different views:
 * - Login and authentication flow
 * - Project selection interface
 * - Presentation generation and viewing
 * - OAuth callback handling
 */
app.use(router)

/**
 * Mount the application to the DOM element with id 'app'.
 * This starts the Vue application lifecycle and renders the initial component tree.
 */
app.mount('#app')