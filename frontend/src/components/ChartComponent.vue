<!--
/**
 * Template for the Chart component.
 * 
 * Provides a responsive container with a canvas element for Chart.js rendering.
 * The canvas uses a template ref and unique ID for proper Chart.js initialization.
 */
-->
<template>
  <div class="chart-component">
    <!-- Canvas element for Chart.js rendering with unique ID and template ref -->
    <canvas ref="chartCanvas" :id="chartId"></canvas>
  </div>
</template>

<!--
/**
 * @fileoverview Chart component for rendering Chart.js visualizations in slides.
 * 
 * This component provides a reusable wrapper around Chart.js that integrates
 * seamlessly with the Intelligent Presenter slide generation system. It handles
 * chart lifecycle management, responsive rendering, and configuration updates.
 * 
 * Features:
 * - Automatic chart initialization and cleanup
 * - Responsive sizing and aspect ratio management
 * - Dynamic configuration updates with reactivity
 * - Error handling for invalid chart configurations
 * - Unique chart instance management
 * 
 * @author Technical Challenge
 * @version 1.0.0
 */
-->

<script setup lang="ts">
import { ref, onMounted, onUnmounted, watch } from 'vue'
import { Chart } from 'chart.js'

/**
 * Component props interface for ChartComponent.
 * 
 * @interface Props
 * @property {any} config - Chart.js configuration object (required)
 * @property {number} [width=400] - Optional chart width in pixels
 * @property {number} [height=300] - Optional chart height in pixels
 */
interface Props {
  /** Chart.js configuration object containing chart type, data, and options */
  config: any
  /** Chart width in pixels (default: 400) */
  width?: number
  /** Chart height in pixels (default: 300) */
  height?: number
}

/**
 * Component props with default values.
 * Provides sensible defaults for width and height while making config required.
 */
const props = withDefaults(defineProps<Props>(), {
  width: 400,
  height: 300
})

/** Reference to the canvas element for Chart.js rendering */
const chartCanvas = ref<HTMLCanvasElement>()

/** Unique identifier for this chart instance */
const chartId = ref(`chart-${Math.random().toString(36).substr(2, 9)}`)

/** Chart.js instance for this component */
let chartInstance: Chart | null = null

/**
 * Creates a new Chart.js instance with the provided configuration.
 * 
 * This function:
 * 1. Validates that the canvas element is available
 * 2. Merges component props with Chart.js configuration
 * 3. Sets up responsive behavior and default styling
 * 4. Handles chart creation errors gracefully
 * 
 * @example
 * Chart configuration is merged with defaults for consistent styling:
 * - Responsive: true for automatic resizing
 * - Legend position: top for better readability
 * - Title display: based on configuration presence
 */
const createChart = () => {
  if (!chartCanvas.value) return

  try {
    chartInstance = new Chart(chartCanvas.value, {
      ...props.config,
      options: {
        responsive: true,
        maintainAspectRatio: false,
        plugins: {
          legend: {
            position: 'top' as const,
          },
          title: {
            display: !!props.config.options?.plugins?.title?.text,
            text: props.config.options?.plugins?.title?.text || ''
          }
        },
        ...props.config.options
      }
    })
  } catch (error) {
    console.error('Failed to create chart:', error)
  }
}

/**
 * Destroys the current Chart.js instance and cleans up resources.
 * 
 * This function:
 * 1. Calls Chart.js destroy method to clean up event listeners
 * 2. Removes references to prevent memory leaks
 * 3. Resets the chart instance to null for garbage collection
 * 
 * Called when component unmounts or when chart configuration changes.
 */
const destroyChart = () => {
  if (chartInstance) {
    chartInstance.destroy()
    chartInstance = null
  }
}

/**
 * Watch for changes in chart configuration and recreate chart.
 * Uses deep watching to detect changes in nested configuration objects.
 */
watch(() => props.config, () => {
  destroyChart()
  createChart()
}, { deep: true })

/**
 * Initialize chart when component is mounted to the DOM.
 * Ensures canvas element is available before chart creation.
 */
onMounted(() => {
  createChart()
})

/**
 * Clean up chart resources when component is unmounted.
 * Prevents memory leaks and ensures proper cleanup.
 */
onUnmounted(() => {
  destroyChart()
})
</script>

<!--
/**
 * Scoped styles for the Chart component.
 * 
 * Provides responsive layout and proper sizing for chart containers:
 * - Relative positioning for proper Chart.js rendering
 * - Full width with fixed height for consistent appearance
 * - Margin spacing for slide layout integration
 * - Responsive canvas sizing
 */
-->
<style scoped>
/* Main container for chart with responsive sizing */
.chart-component {
  position: relative;     /* Required for Chart.js positioning */
  width: 100%;           /* Full width of parent container */
  height: 300px;         /* Fixed height for consistent appearance */
  margin: 1rem 0;        /* Vertical spacing in slide layout */
}

/* Canvas element styling for responsive charts */
canvas {
  max-width: 100%;       /* Prevent overflow in containers */
  height: auto;          /* Maintain aspect ratio */
}
</style>