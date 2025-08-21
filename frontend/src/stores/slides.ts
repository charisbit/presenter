/**
 * @fileoverview Pinia store for managing slide presentation state and operations.
 * This store handles slide generation, real-time updates via WebSocket,
 * navigation between slides, and audio/narration management.
 */

import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { slideApi, websocketService } from '@/services/api'
import { logMarkdown } from '@/utils/markdownLogger'
import type { 
  SlideGenerationRequest, 
  SlideContent, 
  SlideNarration, 
  SlideAudio,
  SlideTheme 
} from '@/types/slides'

/**
 * Pinia store for managing presentation slides and related functionality.
 * 
 * This store provides:
 * - Slide generation and management
 * - Real-time WebSocket communication for generation updates
 * - Navigation between slides
 * - Audio and narration handling
 * - Generation progress tracking
 * 
 * @example
 * ```typescript
 * const slidesStore = useSlidesStore()
 * 
 * // Generate slides
 * await slidesStore.generateSlides({
 *   projectId: '123',
 *   themes: ['project_overview', 'project_progress'],
 *   language: 'ja'
 * })
 * 
 * // Navigate slides
 * slidesStore.nextSlide()
 * slidesStore.goToSlide(2)
 * ```
 */
export const useSlidesStore = defineStore('slides', () => {
  const currentSlideId = ref<string | null>(null)
  const slides = ref<SlideContent[]>([])
  const narrations = ref<Map<number, SlideNarration>>(new Map())
  const audioFiles = ref<Map<number, SlideAudio>>(new Map())
  const currentSlideIndex = ref(0)
  const isGenerating = ref(false)
  const isStreamingComplete = ref(false)
  const websocketConnected = ref(false)
  const expectedTotalSlides = ref(10) // Default to 10 slides
  const slideGenerationStatus = ref<Map<number, 'pending' | 'generating' | 'completed'>>(new Map())
  const expectedThemes = ref<SlideTheme[]>([]) // Store the original themes order

  const currentSlide = computed(() => slides.value[currentSlideIndex.value])
  const totalSlides = computed(() => isStreamingComplete.value ? slides.value.length : expectedTotalSlides.value)
  const hasSlides = computed(() => slides.value.length > 0)
  const canStartPresentation = computed(() => hasSlides.value || isGenerating.value)
  const currentSlideStatus = computed(() => {
    const status = currentSlideIndex.value < slides.value.length 
      ? 'completed' 
      : slideGenerationStatus.value.get(currentSlideIndex.value) || 'pending'
    return status
  })
  const isCurrentSlideGenerating = computed(() => currentSlideStatus.value === 'generating')
  const isCurrentSlideReady = computed(() => currentSlideStatus.value === 'completed')

  /**
   * Initiates slide generation for a Backlog project with specified themes.
   * 
   * This function starts the asynchronous slide generation process by:
   * 1. Clearing any existing slides and state
   * 2. Making an API request to start generation
   * 3. Establishing WebSocket connection for real-time updates
   * 
   * @param request - Configuration for slide generation
   * @param request.projectId - Backlog project identifier
   * @param request.themes - Array of slide themes to generate
   * @param request.language - Target language ('ja' or 'en')
   * 
   * @returns Promise that resolves to the generation response with slideId and WebSocket URL
   * 
   * @throws {Error} If the API request fails or authentication is invalid
   * 
   * @example
   * ```typescript
   * try {
   *   const response = await generateSlides({
   *     projectId: 'PROJECT_123',
   *     themes: ['project_overview', 'team_collaboration'],
   *     language: 'ja'
   *   })
   *   console.log('Generation started:', response.slideId)
   * } catch (error) {
   *   console.error('Failed to start generation:', error)
   * }
   * ```
   */
  const generateSlides = async (request: SlideGenerationRequest) => {
    isGenerating.value = true
    isStreamingComplete.value = false
    slides.value = []
    narrations.value.clear()
    audioFiles.value.clear()
    slideGenerationStatus.value.clear()
    currentSlideIndex.value = 0
    expectedTotalSlides.value = request.themes?.length || 10
    expectedThemes.value = request.themes || []

    // Initialize slide generation status
    for (let i = 0; i < expectedTotalSlides.value; i++) {
      slideGenerationStatus.value.set(i, 'pending')
    }
    
    // Set first slide to generating immediately
    if (expectedTotalSlides.value > 0) {
      slideGenerationStatus.value.set(0, 'generating')
    }

    try {
      const response = await slideApi.generateSlides(request)
      currentSlideId.value = response.slideId

      // Connect to WebSocket for real-time updates
      connectWebSocket(response.slideId)

      return response
    } catch (error) {
      isGenerating.value = false
      throw error
    }
  }

  const connectWebSocket = (slideId: string) => {
    const token = localStorage.getItem('auth_token')
    if (!token) return

    websocketService.connect(`/ws/slides/${slideId}?token=${token}`, {
      onOpen: () => {
        websocketConnected.value = true
      },
      onClose: () => {
        websocketConnected.value = false
      },
      onMessage: (data) => {
        handleWebSocketMessage(data)
      },
      onError: (error) => {
        console.error('WebSocket error:', error)
        websocketConnected.value = false
      }
    })
  }

  const handleWebSocketMessage = (data: any) => {
    switch (data.type) {
      case 'slide_generation_started':
        if (data.data?.slideIndex !== undefined) {
          slideGenerationStatus.value.set(data.data.slideIndex, 'generating')
        }
        break
      case 'slide_content':
        addSlideContent(data.data)
        break
      case 'slide_narration':
        addSlideNarration(data.data)
        break
      case 'slide_audio':
        addSlideAudio(data.data)
        break
      case 'presentation_complete':
        isGenerating.value = false
        isStreamingComplete.value = true
        break
      case 'error':
        console.error('Slide generation error:', data.data)
        isGenerating.value = false
        break
    }
  }

  const addSlideContent = (slideContent: SlideContent) => {
    // Log the received markdown for debugging
    logMarkdown(
      slideContent.markdown, 
      slides.value.length, 
      slideContent.title
    )
    
    slides.value.push(slideContent)
    
    // Mark this slide as completed
    if (slideContent.index !== undefined) {
      slideGenerationStatus.value.set(slideContent.index, 'completed')
    }
    
    
    // If this is the first slide, we can start presentation
    if (slides.value.length === 1) {
      // Don't set isGenerating to false yet, but allow presentation to start
    }
  }

  const addSlideNarration = (narration: SlideNarration) => {
    narrations.value.set(narration.slideIndex, narration)
  }

  const addSlideAudio = (audio: SlideAudio) => {
    audioFiles.value.set(audio.slideIndex, audio)
  }

  /**
   * Advances to the next slide in the presentation.
   * Can navigate to slides that are being generated or pending.
   * 
   * @example
   * ```typescript
   * slidesStore.nextSlide()
   * ```
   */
  const nextSlide = () => {
    if (currentSlideIndex.value < totalSlides.value - 1) {
      currentSlideIndex.value++
    }
  }

  /**
   * Goes back to the previous slide in the presentation.
   * Does nothing if already on the first slide.
   * 
   * @example
   * ```typescript
   * slidesStore.previousSlide()
   * ```
   */
  const previousSlide = () => {
    if (currentSlideIndex.value > 0) {
      currentSlideIndex.value--
    }
  }

  /**
   * Navigates directly to a specific slide by index.
   * Can navigate to slides that are being generated or pending.
   * 
   * @param index - Zero-based index of the target slide
   * 
   * @example
   * ```typescript
   * slidesStore.goToSlide(2) // Go to third slide
   * ```
   */
  const goToSlide = (index: number) => {
    if (index >= 0 && index < totalSlides.value) {
      currentSlideIndex.value = index
    }
  }

  const getNarration = (slideIndex: number): SlideNarration | undefined => {
    return narrations.value.get(slideIndex)
  }

  const getAudio = (slideIndex: number): SlideAudio | undefined => {
    return audioFiles.value.get(slideIndex)
  }

  /**
   * Loads existing slides data from the server for a given slideId.
   * Used for restoring state after page refresh.
   * 
   * @param slideId - The slide session ID to load
   * @returns Promise that resolves when data is loaded
   */
  const loadSlidesFromServer = async (slideId: string) => {
    try {
      const response = await slideApi.getSlideStatus(slideId)
      
      if (response.slides && response.slides.length > 0) {
        // Restore basic session info
        currentSlideId.value = slideId
        expectedThemes.value = response.themes || []
        expectedTotalSlides.value = response.themes?.length || 10
        isGenerating.value = response.status !== 'completed'
        isStreamingComplete.value = response.status === 'completed'
        
        // Restore slides data
        slides.value = response.slides
        
        // Restore narrations
        if (response.narrations) {
          narrations.value.clear()
          response.narrations.forEach((narration: any) => {
            narrations.value.set(narration.slideIndex, narration)
          })
        }
        
        // Restore audio files
        if (response.audioFiles) {
          audioFiles.value.clear()
          response.audioFiles.forEach((audio: any) => {
            audioFiles.value.set(audio.slideIndex, audio)
          })
        }
        
        // Set up slide generation status
        slideGenerationStatus.value.clear()
        for (let i = 0; i < expectedTotalSlides.value; i++) {
          if (i < slides.value.length) {
            slideGenerationStatus.value.set(i, 'completed')
          } else if (isGenerating.value && i === slides.value.length) {
            slideGenerationStatus.value.set(i, 'generating')
          } else {
            slideGenerationStatus.value.set(i, 'pending')
          }
        }
        
        console.log(`Loaded ${slides.value.length} slides from server`)
        return true
      }
      
      return false
    } catch (error) {
      console.error('Failed to load slides from server:', error)
      return false
    }
  }

  const clearSlides = () => {
    currentSlideId.value = null
    slides.value = []
    narrations.value.clear()
    audioFiles.value.clear()
    slideGenerationStatus.value.clear()
    currentSlideIndex.value = 0
    isGenerating.value = false
    isStreamingComplete.value = false
    expectedTotalSlides.value = 10
    expectedThemes.value = []
    websocketService.disconnect()
    websocketConnected.value = false
  }

  return {
    currentSlideId,
    slides,
    currentSlide,
    currentSlideIndex,
    totalSlides,
    hasSlides,
    isGenerating,
    isStreamingComplete,
    websocketConnected,
    canStartPresentation,
    currentSlideStatus,
    isCurrentSlideGenerating,
    isCurrentSlideReady,
    expectedTotalSlides,
    slideGenerationStatus,
    expectedThemes,
    generateSlides,
    connectWebSocket,
    nextSlide,
    previousSlide,
    goToSlide,
    getNarration,
    getAudio,
    loadSlidesFromServer,
    clearSlides
  }
})