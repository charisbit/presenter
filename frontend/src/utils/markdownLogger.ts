/**
 * Simple Markdown logging utility for debugging
 */

/**
 * Log received markdown content to console
 * @param markdown - Raw markdown content
 * @param slideIndex - Index of the slide (optional)
 * @param slideTitle - Title of the slide (optional)
 */
export function logMarkdown(markdown: string, slideIndex?: number, slideTitle?: string): void {
  // Always include function in build, check dev mode at runtime
  if (import.meta.env.DEV || import.meta.env.VITE_ENABLE_MARKDOWN_LOGGING === 'true') {
    // Type check to ensure markdown is a string
    if (typeof markdown !== 'string') {
      console.warn('logMarkdown: Expected string, got:', typeof markdown)
      return
    }

    const prefix = slideIndex !== undefined ? `[Slide ${slideIndex}]` : '[Markdown]'
    const title = slideTitle ? ` ${slideTitle}` : ''
    
    console.group(`📝 ${prefix}${title}`)
    console.log('Raw Markdown:')
    console.log(markdown)
    console.log('---')
    console.log(`Length: ${markdown.length} characters`)
    console.log(`Lines: ${markdown.split('\n').length}`)
    console.groupEnd()
  }
}

/**
 * Log markdown with a custom label
 * @param label - Custom label for the log
 * @param markdown - Raw markdown content
 */
export function logMarkdownWithLabel(label: string, markdown: string): void {
  // Always include function in build, check dev mode at runtime
  if (import.meta.env.DEV || import.meta.env.VITE_ENABLE_MARKDOWN_LOGGING === 'true') {
    // Type check to ensure markdown is a string
    if (typeof markdown !== 'string') {
      console.warn('logMarkdownWithLabel: Expected string, got:', typeof markdown)
      return
    }

    console.group(`📝 ${label}`)
    console.log(markdown)
    console.groupEnd()
  }
}
