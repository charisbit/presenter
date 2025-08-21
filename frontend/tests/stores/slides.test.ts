/**
 * Test suite for slides store
 * Tests slide generation, navigation, and WebSocket functionality
 */

import { describe, it, expect, beforeEach, vi } from 'vitest'
import { setActivePinia, createPinia } from 'pinia'
import { useSlidesStore } from '@/stores/slides'

// Mock the API services
vi.mock('@/services/api', () => ({
  slideApi: {
    generateSlides: vi.fn(),
    getSlide: vi.fn(),
    getNarration: vi.fn(),
    getAudio: vi.fn(),
  },
  websocketService: {
    connect: vi.fn(),
    disconnect: vi.fn(),
    onSlideGenerated: vi.fn(),
    onGenerationProgress: vi.fn(),
    onGenerationComplete: vi.fn(),
  },
}))

describe('Slides Store', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    vi.clearAllMocks()
    localStorage.clear()
  })

  it('should initialize with default state', () => {
    const slidesStore = useSlidesStore()
    
    expect(slidesStore.slides).toEqual([])
    expect(slidesStore.currentSlideIndex).toBe(0)
    expect(slidesStore.isGenerating).toBe(false)
    expect(slidesStore.currentSlideId).toBeNull()
    expect(slidesStore.totalSlides).toBe(0)
    expect(slidesStore.hasSlides).toBe(false)
    expect(slidesStore.generationProgress).toBe(0)
    expect(slidesStore.websocketConnected).toBe(false)
  })

  it('should handle slide generation request', async () => {
    const slidesStore = useSlidesStore()
    const mockRequest = {
      projectId: '123',
      themes: ['project_overview', 'project_progress'],
      language: 'ja',
    }
    
    const mockResponse = {
      slideId: 'slide-123',
      status: 'started',
    }
    
    const { slideApi } = await import('@/services/api')
    vi.mocked(slideApi.generateSlides).mockResolvedValue(mockResponse)
    
    const result = await slidesStore.generateSlides(mockRequest)
    
    expect(slideApi.generateSlides).toHaveBeenCalledWith(mockRequest)
    expect(slidesStore.isGenerating).toBe(true)
    expect(slidesStore.currentSlideId).toBe(mockResponse.slideId)
    expect(result).toEqual(mockResponse)
  })

  it('should handle slide navigation', () => {
    const slidesStore = useSlidesStore()
    
    // Set up mock slides directly
    slidesStore.slides.push(
      { id: '1', title: 'Slide 1', content: 'Content 1', theme: 'project_overview' },
      { id: '2', title: 'Slide 2', content: 'Content 2', theme: 'project_progress' },
      { id: '3', title: 'Slide 3', content: 'Content 3', theme: 'issue_management' }
    )
    
    expect(slidesStore.totalSlides).toBe(3)
    expect(slidesStore.hasSlides).toBe(true)
    
    // Test next slide
    slidesStore.nextSlide()
    expect(slidesStore.currentSlideIndex).toBe(1)
    
    // Test previous slide
    slidesStore.previousSlide()
    expect(slidesStore.currentSlideIndex).toBe(0)
    
    // Test go to specific slide
    slidesStore.goToSlide(2)
    expect(slidesStore.currentSlideIndex).toBe(2)
  })

  it('should handle navigation boundaries', () => {
    const slidesStore = useSlidesStore()
    
    slidesStore.slides.push(
      { id: '1', title: 'Slide 1', content: 'Content 1', theme: 'project_overview' },
      { id: '2', title: 'Slide 2', content: 'Content 2', theme: 'project_progress' }
    )
    slidesStore.currentSlideIndex = 0
    
    // Test previous at beginning
    slidesStore.previousSlide()
    expect(slidesStore.currentSlideIndex).toBe(0) // Should stay at 0
    
    // Go to last slide
    slidesStore.goToSlide(1)
    expect(slidesStore.currentSlideIndex).toBe(1)
    
    // Test next at end
    slidesStore.nextSlide()
    expect(slidesStore.currentSlideIndex).toBe(1) // Should stay at last slide
  })

  it('should handle invalid navigation', () => {
    const slidesStore = useSlidesStore()
    
    slidesStore.slides.push(
      { id: '1', title: 'Slide 1', content: 'Content 1', theme: 'project_overview' }
    )
    
    // Test invalid indices
    slidesStore.goToSlide(-1)
    expect(slidesStore.currentSlideIndex).toBe(0) // Should stay at current
    
    slidesStore.goToSlide(10)
    expect(slidesStore.currentSlideIndex).toBe(0) // Should stay at current
  })

  it('should compute current slide correctly', () => {
    const slidesStore = useSlidesStore()
    
    const mockSlides = [
      { id: '1', title: 'Slide 1', content: 'Content 1', theme: 'project_overview' },
      { id: '2', title: 'Slide 2', content: 'Content 2', theme: 'project_progress' },
    ]
    
    slidesStore.slides.push(...mockSlides)
    slidesStore.currentSlideIndex = 1
    
    expect(slidesStore.currentSlide).toEqual(mockSlides[1])
  })

  it('should handle generation error states', async () => {
    const slidesStore = useSlidesStore()
    
    const { slideApi } = await import('@/services/api')
    vi.mocked(slideApi.generateSlides).mockRejectedValue(new Error('Generation failed'))
    
    const request = {
      projectId: '123',
      themes: ['project_overview'],
      language: 'ja',
    }
    
    await expect(slidesStore.generateSlides(request)).rejects.toThrow('Generation failed')
    expect(slidesStore.isGenerating).toBe(false)
  })

  it('should clear state on new generation', async () => {
    const slidesStore = useSlidesStore()
    
    // Set up existing state
    slidesStore.slides.push(
      { id: '1', title: 'Old Slide', content: 'Old Content', theme: 'project_overview' }
    )
    slidesStore.currentSlideIndex = 1
    slidesStore.generationProgress = 50
    
    const { slideApi } = await import('@/services/api')
    vi.mocked(slideApi.generateSlides).mockResolvedValue({
      slideId: 'new-slide-123',
      status: 'started',
    })
    
    await slidesStore.generateSlides({
      projectId: '456',
      themes: ['project_progress'],
      language: 'en',
    })
    
    expect(slidesStore.slides).toEqual([])
    expect(slidesStore.currentSlideIndex).toBe(0)
    expect(slidesStore.generationProgress).toBe(0)
  })

  it('should handle narration and audio retrieval', () => {
    const slidesStore = useSlidesStore()
    
    // Test non-existent indices (should return undefined)
    expect(slidesStore.getNarration(99)).toBeUndefined()
    expect(slidesStore.getAudio(99)).toBeUndefined()
    
    // Test with existing index (would need to be set via internal methods)
    expect(slidesStore.getNarration(0)).toBeUndefined()
    expect(slidesStore.getAudio(0)).toBeUndefined()
  })

  it('should handle clearSlides method', () => {
    const slidesStore = useSlidesStore()
    
    // Set up some state
    slidesStore.slides.push(
      { id: '1', title: 'Test Slide', content: 'Test Content', theme: 'project_overview' }
    )
    slidesStore.currentSlideIndex = 1
    slidesStore.currentSlideId = 'test-id'
    slidesStore.generationProgress = 50
    
    slidesStore.clearSlides()
    
    expect(slidesStore.slides).toEqual([])
    expect(slidesStore.currentSlideIndex).toBe(0)
    expect(slidesStore.currentSlideId).toBeNull()
    expect(slidesStore.isGenerating).toBe(false)
    expect(slidesStore.generationProgress).toBe(0)
    expect(slidesStore.websocketConnected).toBe(false)
  })

  it('should handle WebSocket connection state', () => {
    const slidesStore = useSlidesStore()
    
    expect(slidesStore.websocketConnected).toBe(false)
    
    // Simulate connection state change (would normally be done internally)
    slidesStore.websocketConnected = true
    expect(slidesStore.websocketConnected).toBe(true)
  })

  it('should handle generation progress updates', () => {
    const slidesStore = useSlidesStore()
    
    expect(slidesStore.generationProgress).toBe(0)
    
    // Simulate progress update
    slidesStore.generationProgress = 25
    expect(slidesStore.generationProgress).toBe(25)
    
    slidesStore.generationProgress = 100
    expect(slidesStore.generationProgress).toBe(100)
  })

  it('should handle slide ID updates', () => {
    const slidesStore = useSlidesStore()
    
    expect(slidesStore.currentSlideId).toBeNull()
    
    slidesStore.currentSlideId = 'slide-123'
    expect(slidesStore.currentSlideId).toBe('slide-123')
    
    slidesStore.currentSlideId = null
    expect(slidesStore.currentSlideId).toBeNull()
  })

  it('should handle generation state changes', () => {
    const slidesStore = useSlidesStore()
    
    expect(slidesStore.isGenerating).toBe(false)
    
    slidesStore.isGenerating = true
    expect(slidesStore.isGenerating).toBe(true)
    
    slidesStore.isGenerating = false
    expect(slidesStore.isGenerating).toBe(false)
  })

  it('should handle multiple slide operations', () => {
    const slidesStore = useSlidesStore()
    
    const slides = [
      { id: '1', title: 'Slide 1', content: 'Content 1', theme: 'project_overview' },
      { id: '2', title: 'Slide 2', content: 'Content 2', theme: 'project_progress' },
      { id: '3', title: 'Slide 3', content: 'Content 3', theme: 'issue_management' },
    ]
    
    // Add slides manually
    slidesStore.slides.push(...slides)
    
    expect(slidesStore.slides).toHaveLength(3)
    expect(slidesStore.totalSlides).toBe(3)
    expect(slidesStore.hasSlides).toBe(true)
    
    // Navigate through slides
    expect(slidesStore.currentSlideIndex).toBe(0)
    
    slidesStore.nextSlide()
    expect(slidesStore.currentSlideIndex).toBe(1)
    
    slidesStore.nextSlide()
    expect(slidesStore.currentSlideIndex).toBe(2)
    
    // Try to go beyond last slide
    slidesStore.nextSlide()
    expect(slidesStore.currentSlideIndex).toBe(2) // Should stay at last slide
  })

  it('should handle empty slides array', () => {
    const slidesStore = useSlidesStore()
    
    expect(slidesStore.currentSlide).toBeUndefined()
    expect(slidesStore.totalSlides).toBe(0)
    expect(slidesStore.hasSlides).toBe(false)
    
    // Navigation should not crash
    slidesStore.nextSlide()
    slidesStore.previousSlide()
    slidesStore.goToSlide(5)
    
    expect(slidesStore.currentSlideIndex).toBe(0)
  })
})
