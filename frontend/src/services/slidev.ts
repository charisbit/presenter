/**
 * @fileoverview Slidev service for single slide processing using native Slidev parser.
 * 
 * This module provides Slidev-powered functionality for:
 * - Processing individual slide markdown content with Slidev parser
 * - Handling Mermaid diagrams natively
 * - Converting Chart.js configurations
 * - Native Slidev markdown to HTML conversion
 * 
 * @author Technical Challenge
 * @version 3.0.0
 */

import { parseSlide, detectFeatures } from '@slidev/parser'
import type { SlideContent } from '@/types/slides'

/**
 * Native Slidev processor for single slide content.
 * 
 * This class uses Slidev's native parser and processing pipeline
 * for accurate rendering of Slidev-flavored markdown.
 */
export class SlidevProcessor {
  constructor() {
    // No configuration needed - Slidev parser handles everything natively
  }
  /**
   * Processes a single slide's markdown content using Slidev parser.
   * 
   * @param slideContent - The slide content to process
   * @returns Processed slide data with enhanced metadata
   */
  async processSlide(slideContent: SlideContent): Promise<any> {
    if (!slideContent.markdown) {
      return {
        content: '',
        raw: '',
        title: slideContent.title || '',
        level: 1
      }
    }
    
    try {
      // Use Slidev's native parseSlide function
      const parsedSlide = parseSlide(slideContent.markdown)
      
      // Detect features used in the slide
      const features = detectFeatures(slideContent.markdown)
      
      // Add our custom processing for Mermaid fixes and Chart.js
      let processedContent = parsedSlide.content
      processedContent = this.fixMermaidSyntax(processedContent)
      processedContent = this.processChartConfigurations(processedContent)
      
      return {
        ...parsedSlide,
        content: processedContent,
        features,
        title: slideContent.title || parsedSlide.title || '',
      }
    } catch (error) {
      console.error('Failed to parse slide with Slidev:', error)
      return {
        content: slideContent.markdown,
        raw: slideContent.markdown,
        title: slideContent.title || '',
        level: 1
      }
    }
  }

  /**
   * Fixes common Mermaid syntax issues in parsed content.
   * 
   * @param content - Content from Slidev parser
   * @returns Content with fixed Mermaid syntax
   */
  private fixMermaidSyntax(content: string): string {
    // Fix pie chart syntax issues - remove % from pie chart values
    // Mermaid expects numbers, not percentages
    const mermaidRegex = /```mermaid\n([\s\S]*?)\n```/g
    
    return content.replace(mermaidRegex, (match, diagramCode) => {
      let fixedCode = diagramCode.trim()
      
      // Fix pie chart syntax issues
      if (fixedCode.includes('pie')) {
        // Remove % from pie chart values (Mermaid expects numbers, not percentages)
        fixedCode = fixedCode.replace(/:\s*(\d+)%/g, ' : $1')
      }
      
      // Return the fixed Mermaid code block
      return `\`\`\`mermaid\n${fixedCode}\n\`\`\``
    })
  }

  /**
   * Processes Chart.js configurations in markdown content.
   * 
   * Converts Chart.js configuration comments to Vue components.
   * 
   * @param markdown - Input markdown with chart configurations
   * @returns Processed markdown with Vue chart components
   */
  private processChartConfigurations(markdown: string): string {
    const chartRegex = /<!-- CHART: (.*?) -->/g
    
    return markdown.replace(chartRegex, (match, chartData) => {
      try {
        const chartConfig = JSON.parse(chartData)
        return `<ChartComponent :config='${JSON.stringify(chartConfig)}' />`
      } catch (error) {
        console.warn('Failed to parse chart configuration:', error)
        return `<div class="chart-error">📊 チャート設定の解析に失敗しました</div>`
      }
    })
  }

  /**
   * Converts Slidev-parsed slide data to HTML.
   * 
   * @param slideData - Processed slide data from Slidev parser
   * @returns HTML string ready for rendering
   */
  convertToHTML(slideData: any): string {
    if (!slideData) {
      return '<p>コンテンツが見つかりません</p>'
    }
    
    try {
      // If slideData is a string (fallback), return as-is with basic markdown processing
      if (typeof slideData === 'string') {
        return this.basicMarkdownToHTML(slideData)
      }
      
      // Use the content from Slidev parser
      const content = slideData.content || slideData.raw || ''
      
      if (!content.trim()) {
        return '<p>スライド内容が見つかりません</p>'
      }
      
      return this.basicMarkdownToHTML(content)
    } catch (error) {
      console.error('Failed to convert slide data to HTML:', error)
      return '<p>スライドの変換に失敗しました</p>'
    }
  }

  /**
   * Basic markdown to HTML conversion for fallback scenarios.
   * 
   * @param markdown - Raw markdown content
   * @returns Basic HTML conversion
   */
  private basicMarkdownToHTML(markdown: string): string {
    return markdown
      // Headers
      .replace(/^### (.*$)/gm, '<h3>$1</h3>')
      .replace(/^## (.*$)/gm, '<h2>$1</h2>')
      .replace(/^# (.*$)/gm, '<h1>$1</h1>')
      
      // Code blocks (Mermaid will be handled by the frontend)
      .replace(/```mermaid\n([\s\S]*?)\n```/g, '<div class="mermaid">$1</div>')
      .replace(/```(\w+)?\n([\s\S]*?)\n```/g, '<pre><code class="language-$1">$2</code></pre>')
      
      // Lists
      .replace(/^[\*\-] (.*$)/gm, '<li>$1</li>')
      
      // Bold and italic
      .replace(/\*\*(.*?)\*\*/g, '<strong>$1</strong>')
      .replace(/\*(.*?)\*/g, '<em>$1</em>')
      
      // Links
      .replace(/\[([^\]]+)\]\(([^)]+)\)/g, '<a href="$2">$1</a>')
      
      // Paragraphs
      .split('\n\n')
      .map(paragraph => {
        paragraph = paragraph.trim()
        if (!paragraph) return ''
        
        if (paragraph.startsWith('<')) {
          return paragraph
        }
        
        if (paragraph.includes('<li>')) {
          return '<ul>' + paragraph + '</ul>'
        }
        
        return '<p>' + paragraph.replace(/\n/g, '<br>') + '</p>'
      })
      .filter(p => p.length > 0)
      .join('\n')
  }
}

/**
 * Default processor instance for use throughout the application.
 */
export const slidevProcessor = new SlidevProcessor()
