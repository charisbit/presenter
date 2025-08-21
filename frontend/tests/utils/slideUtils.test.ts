/**
 * Test suite for slide utilities
 * Tests markdown processing, Mermaid integration, and slide formatting
 */

import { describe, it, expect, beforeEach, vi } from 'vitest'
import { processBareMessageCode, processChartJSConfigs } from '@/utils/slideUtils'

describe('Slide Utils', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  describe('Bare Mermaid Code Processing', () => {
    it('should wrap bare Mermaid graph code', () => {
      const input = `
# Project Flow

graph TD
  A[Start] --> B[Process]
  B --> C[End]

## Summary
This shows our workflow.
      `
      
      const result = processBareMessageCode(input)
      
      expect(result).toContain('```mermaid')
      expect(result).toContain('graph TD')
      expect(result).toContain('```')
    })

    it('should not double-wrap already wrapped Mermaid code', () => {
      const input = `
# Project Flow

\`\`\`mermaid
graph TD
  A[Start] --> B[Process]
  B --> C[End]
\`\`\`

## Summary
This shows our workflow.
      `
      
      const result = processBareMessageCode(input)
      
      // Should return unchanged since it's already wrapped
      expect(result).toBe(input)
    })

    it('should handle pie chart Mermaid code', () => {
      const input = `
# Statistics

pie title Project Status
  "Completed" : 60
  "In Progress" : 30
  "Pending" : 10

## Analysis
The project is mostly complete.
      `
      
      const result = processBareMessageCode(input)
      
      expect(result).toContain('```mermaid')
      expect(result).toContain('pie title')
    })

    it('should handle Gantt chart Mermaid code', () => {
      const input = `
# Timeline

gantt
  title Project Timeline
  section Development
  Task 1 :done, 2024-01-01, 2024-01-15
  Task 2 :active, 2024-01-16, 2024-01-30

## Schedule
Our development timeline.
      `
      
      const result = processBareMessageCode(input)
      
      expect(result).toContain('```mermaid')
      expect(result).toContain('gantt')
    })

    it('should handle flowchart Mermaid code', () => {
      const input = `
# Process Flow

flowchart TD
  A[Start] --> B{Decision}
  B -->|Yes| C[Action 1]
  B -->|No| D[Action 2]

## Process Description
This flowchart shows our decision process.
      `
      
      const result = processBareMessageCode(input)
      
      expect(result).toContain('```mermaid')
      expect(result).toContain('flowchart TD')
    })

    it('should preserve non-Mermaid content', () => {
      const input = `
# Regular Content

This is just regular markdown content.

- Item 1
- Item 2
- Item 3

## Code Example

\`\`\`javascript
console.log('Hello World');
\`\`\`
      `
      
      const result = processBareMessageCode(input)
      
      // Should not add mermaid blocks to regular content
      expect(result).not.toContain('```mermaid')
      expect(result).toBe(input)
    })

    it('should handle multiple Mermaid diagrams', () => {
      const input = `
# Multiple Diagrams

graph LR
  A --> B

## Another Section

pie title Data
  "A" : 50
  "B" : 50

## Summary
Two diagrams above.
      `
      
      const result = processBareMessageCode(input)
      
      // Should wrap both diagrams
      const mermaidBlocks = (result.match(/```mermaid/g) || []).length
      expect(mermaidBlocks).toBeGreaterThanOrEqual(2)
    })

    it('should handle empty input', () => {
      const result = processBareMessageCode('')
      expect(result).toBe('')
    })

    it('should handle input with only whitespace', () => {
      const input = '   \n\n   \t   \n   '
      const result = processBareMessageCode(input)
      expect(result).toBe(input)
    })

    it('should handle mixed wrapped and unwrapped Mermaid', () => {
      const input = `
# Mixed Content

\`\`\`mermaid
graph TD
  A --> B
\`\`\`

And some bare code:

pie title Status
  "Done" : 70
  "Todo" : 30
      `
      
      const result = processBareMessageCode(input)
      
      // Should not modify the already wrapped part
      expect(result).toContain('```mermaid\ngraph TD')
      // Should wrap the bare pie chart
      expect(result).toMatch(/```mermaid[\s\S]*pie title Status/)
    })

    it('should handle Mermaid code with special characters', () => {
      const input = `
# Special Characters

graph TD
  A["Start: 開始"] --> B["Process: 処理"]
  B --> C["End: 終了"]
      `
      
      const result = processBareMessageCode(input)
      
      expect(result).toContain('```mermaid')
      expect(result).toContain('開始')
      expect(result).toContain('処理')
      expect(result).toContain('終了')
    })

    it('should handle malformed Mermaid syntax (current behavior)', () => {
      const input = `
# Malformed

graph
  A --> 
  --> B

## End
      `
      
      const result = processBareMessageCode(input)
      
      // Current implementation may not wrap incomplete syntax
      // This test documents the actual behavior
      expect(result).toBe(input)
    })
  })

  describe('Chart.js Configuration Processing', () => {
    it('should process Chart.js JSON configurations', () => {
      const input = `
# Sales Data

\`\`\`json
{
  "type": "pie",
  "data": {
    "labels": ["Q1", "Q2", "Q3", "Q4"],
    "datasets": [{
      "data": [25, 30, 20, 25]
    }]
  }
}
\`\`\`

## Analysis
The chart shows quarterly distribution.
      `
      
      const result = processChartJSConfigs(input)
      
      expect(result).toContain('<div class="chart-placeholder"')
      expect(result).toContain('data-chart-config=')
      expect(result).not.toContain('```json')
    })

    it('should handle multiple Chart.js configurations', () => {
      const input = `
# Multiple Charts

\`\`\`json
{
  "type": "bar",
  "data": {
    "labels": ["A", "B"],
    "datasets": [{"data": [1, 2]}]
  }
}
\`\`\`

\`\`\`json
{
  "type": "line",
  "data": {
    "labels": ["X", "Y"],
    "datasets": [{"data": [3, 4]}]
  }
}
\`\`\`
      `
      
      const result = processChartJSConfigs(input)
      
      const chartPlaceholders = (result.match(/chart-placeholder/g) || []).length
      expect(chartPlaceholders).toBe(2)
    })

    it('should ignore JSON blocks without type property', () => {
      const input = `
# Regular JSON

\`\`\`json
{
  "name": "test",
  "value": 123
}
\`\`\`
      `
      
      const result = processChartJSConfigs(input)
      
      // Should remain unchanged
      expect(result).toBe(input)
      expect(result).not.toContain('chart-placeholder')
    })

    it('should handle invalid JSON gracefully', () => {
      // Suppress console.warn for this test since we're testing error handling
      const consoleSpy = vi.spyOn(console, 'warn').mockImplementation(() => {})
      
      const input = `
# Invalid JSON

\`\`\`json
{
  "type": "bar"
  invalid json here
}
\`\`\`
      `
      
      const result = processChartJSConfigs(input)
      
      // Should remain unchanged due to invalid JSON
      expect(result).toBe(input)
      
      // Verify that warning was logged
      expect(consoleSpy).toHaveBeenCalledWith(
        'Invalid Chart.js configuration found:',
        expect.any(SyntaxError)
      )
      
      consoleSpy.mockRestore()
    })

    it('should preserve other code blocks', () => {
      const input = `
# Mixed Code

\`\`\`javascript
console.log('hello');
\`\`\`

\`\`\`json
{
  "type": "pie",
  "data": {"labels": ["A"], "datasets": [{"data": [1]}]}
}
\`\`\`

\`\`\`python
print("world")
\`\`\`
      `
      
      const result = processChartJSConfigs(input)
      
      expect(result).toContain('```javascript')
      expect(result).toContain('```python')
      expect(result).toContain('chart-placeholder')
    })
  })

  describe('Edge Cases and Error Handling', () => {
    it('should handle empty markdown', () => {
      expect(processBareMessageCode('')).toBe('')
      expect(processChartJSConfigs('')).toBe('')
    })

    it('should handle markdown with only whitespace', () => {
      const whitespace = '   \n\t  \n   '
      expect(processBareMessageCode(whitespace)).toBe(whitespace)
      expect(processChartJSConfigs(whitespace)).toBe(whitespace)
    })

    it('should handle very long content', () => {
      const longContent = 'A'.repeat(1000) + '\n\ngraph TD\nA --> B\n\n' + 'B'.repeat(1000)
      const result = processBareMessageCode(longContent)
      
      expect(result).toContain('```mermaid')
      expect(result.length).toBeGreaterThan(longContent.length)
    })

    it('should handle nested code blocks', () => {
      const input = `
# Nested Example

\`\`\`markdown
# Example
\`\`\`mermaid
graph TD
A --> B
\`\`\`
\`\`\`
      `
      
      const result = processBareMessageCode(input)
      
      // Should not process Mermaid inside other code blocks
      expect(result).toBe(input)
    })

    it('should handle content with no Mermaid patterns', () => {
      const input = `
# Regular Markdown

This is just regular content with no diagrams.

## Lists
- Item 1
- Item 2

## Code
\`\`\`javascript
console.log('hello');
\`\`\`
      `
      
      const result = processBareMessageCode(input)
      expect(result).toBe(input)
    })

    it('should handle Chart.js config with nested objects', () => {
      const input = `
# Complex Chart

\`\`\`json
{
  "type": "line",
  "data": {
    "labels": ["Jan", "Feb"],
    "datasets": [{
      "label": "Sales",
      "data": [100, 200],
      "backgroundColor": "rgba(255, 99, 132, 0.2)",
      "borderColor": "rgba(255, 99, 132, 1)"
    }]
  },
  "options": {
    "responsive": true,
    "plugins": {
      "legend": {
        "position": "top"
      }
    }
  }
}
\`\`\`
      `
      
      const result = processChartJSConfigs(input)
      
      expect(result).toContain('chart-placeholder')
      expect(result).toContain('data-chart-config=')
    })
  })
})
