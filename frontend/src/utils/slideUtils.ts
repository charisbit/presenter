/**
 * @fileoverview Utility functions for slide processing and markdown enhancement.
 * 
 * This module provides utility functions for processing slide content,
 * including Mermaid diagram detection and wrapping, Chart.js configuration
 * processing, and theme localization for the Japanese interface.
 * 
 * These utilities are used throughout the slide generation and presentation
 * pipeline to ensure proper rendering of complex content types.
 * 
 * @author Technical Challenge
 * @version 1.0.0
 */

import { logMarkdownWithLabel } from './markdownLogger'

/**
 * Processes markdown content to wrap bare Mermaid code with proper syntax highlighting.
 * 
 * This function detects Mermaid diagram code that isn't already wrapped in 
 * code blocks and adds the proper ```mermaid syntax for rendering. It handles
 * common Mermaid diagram types including graphs, pie charts, Gantt charts,
 * and flowcharts.
 * 
 * @param {string} markdown - The input markdown content potentially containing bare Mermaid code
 * @returns {string} Processed markdown with properly wrapped Mermaid code blocks
 * 
 * @example
 * ```typescript
 * const input = `
 * # Project Flow
 * 
 * graph TD
 *   A[Start] --> B[Process]
 *   B --> C[End]
 * 
 * ## Summary
 * This shows our workflow.
 * `
 * 
 * const output = processBareMessageCode(input)
 * // Result will have the graph wrapped in ```mermaid blocks
 * ```
 */
export const processBareMessageCode = (markdown: string): string => {
  // Log input markdown for debugging
  logMarkdownWithLabel('processBareMessageCode - Input', markdown)
  
  // Check if already wrapped to prevent double wrapping
  if (markdown.includes('```mermaid')) {
    return markdown
  }

  /**
   * Regular expressions for detecting different types of Mermaid diagrams.
   * Each pattern captures:
   * - Line start or newline character
   * - Mermaid diagram type keyword and content
   * - Stops at double newline, heading, or end of string
   */
  const mermaidPatterns = [
    /(^|\n)(graph\s+(?:TD|TB|BT|RL|LR)[\s\S]*?)(?=\n\n|\n#|$)/g,
    /(^|\n)(pie\s+title[\s\S]*?)(?=\n\n|\n#|$)/g,
    /(^|\n)(gantt[\s\S]*?)(?=\n\n|\n#|$)/g,
    /(^|\n)(flowchart\s+(?:TD|TB|BT|RL|LR)[\s\S]*?)(?=\n\n|\n#|$)/g
  ]
  
  let processedMarkdown = markdown
  
  // Apply each pattern to wrap detected Mermaid code
  mermaidPatterns.forEach(pattern => {
    processedMarkdown = processedMarkdown.replace(pattern, (match, prefix, mermaidCode) => {
      return `${prefix}\`\`\`mermaid\n${mermaidCode.trim()}\n\`\`\``
    })
  })
  
  // Log output if changed
  if (processedMarkdown !== markdown) {
    logMarkdownWithLabel('processBareMessageCode - Output', processedMarkdown)
  }
  
  return processedMarkdown
}

/**
 * Processes markdown content to convert Chart.js JSON configurations to HTML placeholders.
 * 
 * This function searches for JSON code blocks containing Chart.js configuration
 * objects and converts them to HTML div elements with data attributes. These
 * placeholders are later processed by the frontend to render actual charts.
 * 
 * @param {string} markdown - The input markdown content containing Chart.js configurations
 * @returns {string} Processed markdown with Chart.js HTML placeholders
 * 
 * @example
 * ```typescript
 * const input = `
 * # Sales Data
 * 
 * ```json
 * {
 *   "type": "pie",
 *   "data": {
 *     "labels": ["Q1", "Q2", "Q3", "Q4"],
 *     "datasets": [{
 *       "data": [25, 30, 20, 25]
 *     }]
 *   }
 * }
 * ```
 * `
 * 
 * const output = processChartJSConfigs(input)
 * // Result will have a <div class="chart-placeholder"> element
 * ```
 */
export const processChartJSConfigs = (markdown: string): string => {
  // Log input markdown for debugging
  logMarkdownWithLabel('processChartJSConfigs - Input', markdown)
  
  /** Regular expression to match JSON code blocks containing Chart.js configs */
  const chartRegex = /```json\s*(\{[\s\S]*?"type"\s*:\s*"[^"]+?"[\s\S]*?\})\s*```/g
  let match
  let processedMarkdown = markdown
  let chartIndex = 0

  // Process each JSON code block that contains a "type" property
  while ((match = chartRegex.exec(markdown)) !== null) {
    try {
      const chartConfig = JSON.parse(match[1])
      if (chartConfig.type) {
        // Generate unique chart identifier
        const chartId = `chart-${Date.now()}-${chartIndex++}`
        
        // Create HTML placeholder with chart configuration
        const chartPlaceholder = `<div class="chart-placeholder" data-chart-config='${JSON.stringify(chartConfig)}' data-chart-id="${chartId}"></div>`
        
        // Replace the JSON code block with the placeholder
        processedMarkdown = processedMarkdown.replace(match[0], chartPlaceholder)
      }
    } catch (error) {
      // Skip invalid JSON configurations
      console.warn('Invalid Chart.js configuration found:', error)
    }
  }

  // Log output if changed
  if (processedMarkdown !== markdown) {
    logMarkdownWithLabel('processChartJSConfigs - Output', processedMarkdown)
  }

  return processedMarkdown
}

/**
 * Gets the localized Japanese label for a slide theme identifier.
 * 
 * This function provides human-readable Japanese labels for the slide themes
 * used throughout the application. It's used in the UI to display friendly
 * names for the different types of slides that can be generated.
 * 
 * @param {string} theme - The theme identifier (e.g., 'project_overview')
 * @returns {string} Localized theme label in Japanese, or the original theme if not found
 * 
 * @example
 * ```typescript
 * const label = getThemeLabel('project_overview')
 * console.log(label) // Output: 'プロジェクト概要'
 * 
 * const unknownLabel = getThemeLabel('unknown_theme')
 * console.log(unknownLabel) // Output: 'unknown_theme'
 * ```
 */
export const getThemeLabel = (theme: string): string => {
  /**
   * Mapping of theme identifiers to Japanese labels.
   * Each theme corresponds to a specific type of slide content
   * generated from Backlog project data.
   */
  const themeLabels: Record<string, string> = {
    'project_overview': 'プロジェクト概要',      // Project overview and basic information
    'project_progress': 'プロジェクト進捗',      // Progress tracking and completion rates
    'issue_management': '課題管理',              // Issue statistics and management
    'risk_analysis': 'リスク分析',               // Risk identification and analysis
    'team_collaboration': 'チーム協力',          // Team activities and collaboration
    'document_management': 'ドキュメント管理',   // Documentation and knowledge sharing
    'codebase_activity': 'コードベース活動',     // Development and code metrics
    'notifications': '通知管理',                 // Communication and notification patterns
    'predictive_analysis': '予測分析',           // Forecasting and trend analysis
    'summary_plan': '総括と計画'                 // Summary and future planning
  }
  
  // Return localized label or fall back to original theme identifier
  return themeLabels[theme] || theme
}