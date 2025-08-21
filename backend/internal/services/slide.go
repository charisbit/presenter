// Package services provides core business logic services for the intelligent presenter application.
// This package includes services for slide generation, content processing, and AI integration.
package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"intelligent-presenter-backend/internal/models"
	"intelligent-presenter-backend/pkg/config"
)

// SlideService provides functionality for generating presentation slides
// using AI-powered content generation and project data from Backlog.
// It integrates with multiple AI providers (OpenAI, AWS Bedrock) and
// supports various slide themes and content types.
type SlideService struct {
	config            *config.Config        // Application configuration
	mcpService        *MCPService          // MCP service for Backlog data access
	bedrockService    *BedrockService      // AWS Bedrock service (custom implementation)
	bedrockSDKService *BedrockSDKService   // AWS Bedrock service (SDK implementation)
}

// NewSlideService creates a new instance of SlideService with the provided configuration.
// It initializes connections to AI services and MCP services for data retrieval.
// The service automatically falls back between different AI providers if one fails.
func NewSlideService(cfg *config.Config) *SlideService {
	// Try to create AWS SDK service, fallback to custom implementation if it fails
	var bedrockSDKService *BedrockSDKService
	if cfg.AWSAccessKeyID != "" && cfg.AWSSecretAccessKey != "" {
		if sdkService, err := NewBedrockSDKService(cfg); err == nil {
			bedrockSDKService = sdkService
		} else {
			fmt.Printf("Failed to create Bedrock SDK service, falling back to custom implementation: %v\n", err)
		}
	}

	return &SlideService{
		config:         cfg,
		mcpService:     NewMCPService(cfg),
		bedrockService: NewBedrockService(cfg),
		bedrockSDKService: bedrockSDKService,
	}
}

// GenerateSlideContent creates a complete slide with both markdown and HTML content
// for the specified project, theme, and language. This is the main entry point
// for slide generation and includes data retrieval, AI content generation,
// and HTML compilation.
//
// Parameters:
//   - projectID: The Backlog project identifier
//   - theme: The slide theme (e.g., project_overview, progress, etc.)
//   - language: Target language for content generation ("ja" or "en")
//   - backlogToken: Authentication token for Backlog API access
//
// Returns:
//   - *models.SlideContent: Complete slide with markdown and HTML content
//   - error: Any error that occurred during generation
func (s *SlideService) GenerateSlideContent(projectID string, theme models.SlideTheme, language, backlogToken string) (*models.SlideContent, error) {
	// Get project data based on theme
	projectData, err := s.getProjectDataForTheme(projectID, theme, backlogToken)
	if err != nil {
		return nil, fmt.Errorf("failed to get project data: %w", err)
	}

	// Generate markdown content using OpenAI
	markdown, title, err := s.generateMarkdownContent(projectData, theme, language)
	if err != nil {
		return nil, fmt.Errorf("failed to generate markdown: %w", err)
	}

	// // Generate HTML from markdown using LLM
	// html, err := s.generateHTMLFromMarkdown(markdown, title, language)
	// if err != nil {
	// 	return nil, fmt.Errorf("failed to generate HTML: %w", err)
	// }

	return &models.SlideContent{
		Theme:       theme,
		Title:       title,
		Markdown:    markdown,
		// HTML:        html,
		GeneratedAt: time.Now(),
	}, nil
}

// GenerateSlideNarration creates spoken narration text for a slide
// using AI-powered natural language generation. The narration is optimized
// for text-to-speech synthesis and presentation delivery.
//
// Parameters:
//   - slide: The slide content to generate narration for
//   - language: Target language for narration ("ja" or "en")
//
// Returns:
//   - *models.SlideNarration: Generated narration with timing information
//   - error: Any error that occurred during generation
func (s *SlideService) GenerateSlideNarration(slide *models.SlideContent, language string) (*models.SlideNarration, error) {
	// Generate narration text using OpenAI
	narrationText, err := s.generateNarrationText(slide.Markdown, slide.Title, language)
	if err != nil {
		return nil, fmt.Errorf("failed to generate narration: %w", err)
	}

	return &models.SlideNarration{
		SlideIndex: slide.Index,
		Text:       narrationText,
		Language:   language,
	}, nil
}

func (s *SlideService) GenerateSlideAudio(narration *models.SlideNarration) (*models.SlideAudio, error) {
	// Use MCP Speech service to synthesize audio
	audioURL, err := s.mcpService.SynthesizeSpeech(narration.Text, narration.Language, "")
	if err != nil {
		return nil, fmt.Errorf("failed to synthesize speech: %w", err)
	}

	// Estimate duration based on text length (rough calculation)
	// Average speaking rate is about 150-160 words per minute
	wordCount := len(strings.Fields(narration.Text))
	if wordCount < 1 {
		wordCount = 1
	}
	duration := (wordCount * 60) / 150 // seconds

	return &models.SlideAudio{
		SlideIndex: narration.SlideIndex,
		AudioURL:   audioURL,
		Duration:   duration,
	}, nil
}

func (s *SlideService) getProjectDataForTheme(projectID string, theme models.SlideTheme, backlogToken string) (map[string]interface{}, error) {
	data := make(map[string]interface{})
	fmt.Printf("Getting project data for theme: %s, projectID: %s\n", theme, projectID)

	switch theme {
	case models.ThemeProjectOverview:
		fmt.Printf("Fetching project overview...\n")
		overview, err := s.mcpService.GetProjectOverview(projectID, backlogToken)
		if err != nil {
			fmt.Printf("Failed to get project overview: %v\n", err)
			return nil, err
		}
		data["overview"] = overview
		fmt.Printf("Project overview fetched successfully\n")

	case models.ThemeProjectProgress:
		fmt.Printf("Fetching project progress...\n")
		progress, err := s.mcpService.GetProjectProgress(projectID, backlogToken)
		if err != nil {
			fmt.Printf("Failed to get project progress: %v\n", err)
			return nil, err
		}
		data["progress"] = progress
		fmt.Printf("Project progress fetched successfully\n")

	case models.ThemeIssueManagement:
		fmt.Printf("Fetching project issues...\n")
		issues, err := s.mcpService.GetProjectIssues(projectID, backlogToken)
		if err != nil {
			fmt.Printf("Failed to get project issues: %v\n", err)
			return nil, err
		}
		data["issues"] = issues
		fmt.Printf("Project issues fetched successfully\n")

	case models.ThemeTeamCollaboration:
		fmt.Printf("Fetching project team...\n")
		team, err := s.mcpService.GetProjectTeam(projectID, backlogToken)
		if err != nil {
			fmt.Printf("Failed to get project team: %v\n", err)
			// For team collaboration, use fallback data when API fails
			fmt.Printf("Using fallback team data for team collaboration slide\n")
			data["team"] = map[string]interface{}{
				"users": []map[string]interface{}{
					{"name": "プロジェクトメンバー", "role": "開発者"},
				},
				"fallback": true,
				"error": "API access limited - using sample data",
			}
		} else {
			data["team"] = team
		}
		fmt.Printf("Project team data prepared successfully\n")

	case models.ThemeRiskAnalysis:
		fmt.Printf("Fetching project risks...\n")
		risks, err := s.mcpService.GetProjectRisks(projectID, backlogToken)
		if err != nil {
			fmt.Printf("Failed to get project risks: %v\n", err)
			return nil, err
		}
		data["risks"] = risks
		fmt.Printf("Project risks fetched successfully\n")

	case models.ThemeDocumentManagement:
		fmt.Printf("Fetching project documents...\n")
		// Get Wiki and document information
		overview, err := s.mcpService.GetProjectOverview(projectID, backlogToken)
		if err != nil {
			fmt.Printf("Failed to get project overview for documents: %v\n", err)
			return nil, err
		}
		data["overview"] = overview
		data["focus"] = "documents"
		fmt.Printf("Project documents fetched successfully\n")

	case models.ThemeCodebaseActivity:
		fmt.Printf("Fetching project codebase activity...\n")
		// Get Git repository and development activity information
		overview, err := s.mcpService.GetProjectOverview(projectID, backlogToken)
		if err != nil {
			fmt.Printf("Failed to get project overview for codebase: %v\n", err)
			return nil, err
		}
		data["overview"] = overview
		data["focus"] = "codebase"
		fmt.Printf("Project codebase activity fetched successfully\n")

	case models.ThemeNotifications:
		fmt.Printf("Fetching project notifications...\n")
		// Get notification and communication information
		overview, err := s.mcpService.GetProjectOverview(projectID, backlogToken)
		if err != nil {
			fmt.Printf("Failed to get project overview for notifications: %v\n", err)
			return nil, err
		}
		data["overview"] = overview
		data["focus"] = "notifications"
		fmt.Printf("Project notifications fetched successfully\n")

	case models.ThemePredictiveAnalysis:
		fmt.Printf("Fetching project data for predictive analysis...\n")
		// Get project progress and issues for predictive analysis
		progress, err := s.mcpService.GetProjectProgress(projectID, backlogToken)
		if err != nil {
			fmt.Printf("Failed to get project progress for prediction: %v\n", err)
			return nil, err
		}
		issues, err2 := s.mcpService.GetProjectIssues(projectID, backlogToken)
		if err2 != nil {
			fmt.Printf("Failed to get project issues for prediction: %v\n", err2)
			return nil, err2
		}
		data["progress"] = progress
		data["issues"] = issues
		data["focus"] = "prediction"
		fmt.Printf("Project data for predictive analysis fetched successfully\n")

	case models.ThemeSummaryPlan:
		fmt.Printf("Fetching comprehensive project data for summary...\n")
		// Get comprehensive data for summary and planning
		overview, err := s.mcpService.GetProjectOverview(projectID, backlogToken)
		if err != nil {
			fmt.Printf("Failed to get project overview for summary: %v\n", err)
			return nil, err
		}
		progress, err2 := s.mcpService.GetProjectProgress(projectID, backlogToken)
		if err2 != nil {
			fmt.Printf("Failed to get project progress for summary: %v\n", err2)
			// Non-critical, continue with overview only
			progress = nil
		}
		data["overview"] = overview
		data["progress"] = progress
		data["focus"] = "summary"
		fmt.Printf("Comprehensive project data for summary fetched successfully\n")

	default:
		fmt.Printf("Using default theme, fetching project overview...\n")
		// For other themes, get general project data
		overview, err := s.mcpService.GetProjectOverview(projectID, backlogToken)
		if err != nil {
			fmt.Printf("Failed to get default project overview: %v\n", err)
			return nil, err
		}
		data["overview"] = overview
		fmt.Printf("Default project overview fetched successfully\n")
	}

	fmt.Printf("Project data collection completed for theme: %s\n", theme)
	return data, nil
}

func (s *SlideService) generateMarkdownContent(projectData map[string]interface{}, theme models.SlideTheme, language string) (string, string, error) {
	prompt := s.buildPromptForTheme(projectData, theme, language)

	// Call AI API based on provider
	var response string
	var err error
	
	fmt.Printf("Using AI provider: %s\n", s.config.AIProvider)
	
	switch s.config.AIProvider {
	case "bedrock":
		response, err = s.callBedrock(prompt)
		// Auto-fallback to OpenAI if Bedrock fails
		if err != nil {
			fmt.Printf("Bedrock API failed: %v, falling back to OpenAI\n", err)
			response, err = s.callOpenAI(prompt)
			if err != nil {
				fmt.Printf("OpenAI fallback also failed: %v\n", err)
				return "", "", err
			}
			fmt.Printf("OpenAI fallback successful\n")
		}
	case "openai":
		response, err = s.callOpenAI(prompt)
	default:
		// Default to OpenAI if not specified
		response, err = s.callOpenAI(prompt)
	}
	
	if err != nil {
		fmt.Printf("AI API call failed: %v\n", err)
		return "", "", err
	}

	// Define theme-specific default titles
	themeDefaultTitles := map[models.SlideTheme]string{
		models.ThemeProjectOverview:     "プロジェクト概要",
		models.ThemeProjectProgress:     "プロジェクト進捗",
		models.ThemeIssueManagement:     "課題管理",
		models.ThemeRiskAnalysis:        "リスク分析",
		models.ThemeTeamCollaboration:   "チーム協力",
		models.ThemeDocumentManagement:  "ドキュメント管理",
		models.ThemeCodebaseActivity:    "コードベース活動",
		models.ThemeNotifications:       "通知管理",
		models.ThemePredictiveAnalysis:  "予測分析",
		models.ThemeSummaryPlan:         "総括と計画",
	}

	themeDefaultTitlesEN := map[models.SlideTheme]string{
		models.ThemeProjectOverview:     "Project Overview",
		models.ThemeProjectProgress:     "Project Progress",
		models.ThemeIssueManagement:     "Issue Management",
		models.ThemeRiskAnalysis:        "Risk Analysis",
		models.ThemeTeamCollaboration:   "Team Collaboration",
		models.ThemeDocumentManagement:  "Document Management",
		models.ThemeCodebaseActivity:    "Codebase Activity",
		models.ThemeNotifications:       "Notifications",
		models.ThemePredictiveAnalysis:  "Predictive Analysis",
		models.ThemeSummaryPlan:         "Summary & Plan",
	}

	// Extract title and markdown from response
	lines := strings.Split(response, "\n")
	
	// Set default title based on theme and language
	var title string
	if language == "ja" {
		if defaultTitle, exists := themeDefaultTitles[theme]; exists {
			title = defaultTitle
		} else {
			title = "Project Slide"
		}
	} else {
		if defaultTitle, exists := themeDefaultTitlesEN[theme]; exists {
			title = defaultTitle
		} else {
			title = "Project Slide"
		}
	}
	
	markdown := response

	// Look for title in first line if it starts with #
	if len(lines) > 0 && strings.HasPrefix(lines[0], "#") {
		extractedTitle := strings.TrimSpace(strings.TrimPrefix(lines[0], "#"))
		fmt.Printf("AI generated title: '%s' for theme: %s\n", extractedTitle, theme)
		title = extractedTitle
	} else {
		fmt.Printf("No # title found, using default title: '%s' for theme: %s\n", title, theme)
		fmt.Printf("First line of AI response: '%s'\n", lines[0])
	}

	return markdown, title, nil
}

func (s *SlideService) generateNarrationText(markdown, title, language string) (string, error) {
	var prompt string
	if language == "ja" {
		prompt = fmt.Sprintf(`
以下のMarkdown形式のスライド内容に基づいて、日本語で自然な口頭発表用のナレーションを生成してください。

スライド内容:
%s

ナレーションの要件:
1. 聞き手に分かりやすい自然な日本語
2. プロフェッショナルなプレゼンテーション調
3. 2-3分程度で読める長さ
4. スライドの内容を効果的に説明

ナレーション:`, markdown)
	} else {
		prompt = fmt.Sprintf(`
Generate natural narration text in English for the following slide content:

Slide Content:
%s

Requirements:
1. Natural, professional presentation style
2. 2-3 minutes reading time
3. Clear explanation of slide content

Narration:`, markdown)
	}

	// Use the same AI provider as for content generation with fallback
	switch s.config.AIProvider {
	case "bedrock":
		response, err := s.callBedrock(prompt)
		// Auto-fallback to OpenAI if Bedrock fails
		if err != nil {
			fmt.Printf("Bedrock narration API failed: %v, falling back to OpenAI\n", err)
			response, err = s.callOpenAI(prompt)
			if err != nil {
				fmt.Printf("OpenAI narration fallback also failed: %v\n", err)
				return "", err
			}
			fmt.Printf("OpenAI narration fallback successful\n")
		}
		return response, err
	case "openai":
		return s.callOpenAI(prompt)
	default:
		return s.callOpenAI(prompt)
	}
}

func (s *SlideService) buildPromptForTheme(projectData map[string]interface{}, theme models.SlideTheme, language string) string {
	// Limit the data size to prevent context overflow
	dataJSON, _ := json.Marshal(projectData)
	if len(dataJSON) > 8000 { // Limit to ~8KB to keep under token limits
		dataJSON = dataJSON[:8000]
		dataJSON = append(dataJSON, []byte("...}")...) // Close JSON properly
	}

	themePrompts := map[models.SlideTheme]string{
		models.ThemeProjectOverview: `プロジェクトの概要と基本情報のスライドを生成してください。プロジェクト名、目的、期間、チーム構成などを含めてください。`,
		models.ThemeProjectProgress: `プロジェクトの進捗状況のスライドを生成してください。完了率、マイルストーン、現在の状況などを含めてください。`,
		models.ThemeIssueManagement: `プロジェクトの課題管理状況のスライドを生成してください。未解決の課題、優先度分布、進行中のタスクなどを含めてください。`,
		models.ThemeRiskAnalysis: `プロジェクトのリスク分析のスライドを生成してください。潜在的なリスク、遅延要因、対策などを含めてください。`,
		models.ThemeTeamCollaboration: `チームの協力状況のスライドを生成してください。メンバー構成、役割分担、コミュニケーション状況などを含めてください。`,
		models.ThemeDocumentManagement: `プロジェクトの文書管理状況のスライドを生成してください。文書数、更新頻度、アクセス状況、知識共有などを含めてください。`,
		models.ThemeCodebaseActivity: `プロジェクトの開発活動のスライドを生成してください。コミット数、開発者活動量、コード品質指標、リリース頻度などを含めてください。`,
		models.ThemeNotifications: `プロジェクトのコミュニケーション状況のスライドを生成してください。通知数、応答率、情報伝達効率、重要通知の処理状況などを含めてください。`,
		models.ThemePredictiveAnalysis: `プロジェクトの予測分析のスライドを生成してください。完了予測日、リスク発生確率、必要リソース予測、目標達成可能性などを含めてください。`,
		models.ThemeSummaryPlan: `プロジェクトの総括・計画のスライドを生成してください。主要成果、KPI達成状況、残課題、次期計画の要点などを含めてください。`,
	}

	themePromptsEN := map[models.SlideTheme]string{
		models.ThemeProjectOverview: "Generate a slide for project overview and basic information. Include project name, purpose, duration, team composition, etc.",
		models.ThemeProjectProgress: "Generate a slide for project progress status. Include completion rate, milestones, current status, etc.",
		models.ThemeIssueManagement: "Generate a slide for project issue management status. Include unresolved issues, priority distribution, ongoing tasks, etc.",
		models.ThemeRiskAnalysis: "Generate a slide for project risk analysis. Include potential risks, delay factors, countermeasures, etc.",
		models.ThemeTeamCollaboration: "Generate a slide for team collaboration status. Include member composition, role assignments, communication status, etc.",
		models.ThemeDocumentManagement: "Generate a slide for project document management status. Include document count, update frequency, access status, knowledge sharing, etc.",
		models.ThemeCodebaseActivity: "Generate a slide for project development activity. Include commit count, developer activity levels, code quality metrics, release frequency, etc.",
		models.ThemeNotifications: "Generate a slide for project communication status. Include notification count, response rate, information transmission efficiency, important notification processing status, etc.",
		models.ThemePredictiveAnalysis: "Generate a slide for project predictive analysis. Include predicted completion date, risk occurrence probability, required resource forecast, goal achievement feasibility, etc.",
		models.ThemeSummaryPlan: "Generate a slide for project summary and planning. Include key achievements, KPI achievement status, remaining issues, key points of next plan, etc.",
	}

	var themePrompt string
	var exists bool

	if language == "ja" {
		themePrompt, exists = themePrompts[theme]
		if !exists {
			themePrompt = "プロジェクト関連のスライドを生成してください。"
		}
		return fmt.Sprintf(`
以下のBacklogプロジェクトデータを基に、%s

データ:
%s

要件:
1. **必ず # で始まるタイトル行から開始してください**
2. **上司への報告用**として簡潔に作成
3. スライドは1枚、レイアウトはコンパクトに、3-5個の要点のみ（詳細は避ける）
4. データ可視化のため以下のうち1つを含める：
   - Mermaidダイアグラム（シンプルなフローチャート、円グラフ、ガントチャートなど）
   - Chart.jsグラフ（必要に応じて）
5. 箇条書きを多用し、読みやすく構成
6. 数値や結果を強調
7. Mermaidを使用する場合は ` + "```" + `mermaid で始めること
8. **重要**: 冗長な説明は避け、核心的な情報のみ記載

スライド内容:`, themePrompt, string(dataJSON))
	} else {
		themePrompt, exists = themePromptsEN[theme]
		if !exists {
			themePrompt = "Generate a slide about the project."
		}
		return fmt.Sprintf(`
Generate a slide based on the following Backlog project data for theme: %s

Data:
%s

Requirements:
1. **Must start with a title line beginning with #**
2. **Executive briefing format** - concise and focused
3. Only generate one slide; use a compact layout.　Maximum 3-5 key points (avoid details)
4. Include one data visualization:
   - Simple Mermaid diagrams (flowcharts, pie charts, gantt charts)
   - Chart.js graphs (when appropriate)
5. Use bullet points for readability
6. Emphasize numbers and results
7. For Mermaid, use ` + "```" + `mermaid code blocks
8. **Important**: Avoid verbose explanations, focus on core information only
9. **Important**: Only generate one slide
10. **Important**: Use a compact layout

Slide Content:`, themePrompt, string(dataJSON))
	}
}

func (s *SlideService) callOpenAI(prompt string) (string, error) {
	if s.config.OpenAIAPIKey == "" {
		return "", fmt.Errorf("OpenAI API key not configured")
	}

	requestBody := map[string]interface{}{
		"model": "gpt-3.5-turbo",
		"messages": []map[string]string{
			{
				"role":    "user",
				"content": prompt,
			},
		},
		"max_tokens":  800, // Reduced to prevent context overflow
		"temperature": 0.7,
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Printf("OpenAI request marshal error: %v\n", err)
		return "", err
	}

	req, err := http.NewRequest("POST", "https://api.openai.com/v1/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Printf("OpenAI request creation error: %v\n", err)
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.config.OpenAIAPIKey)

	fmt.Printf("Making OpenAI API call...\n")
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("OpenAI API call error: %v\n", err)
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		fmt.Printf("OpenAI API error - Status: %d\n", resp.StatusCode)
		// Read error response
		var errorBytes bytes.Buffer
		errorBytes.ReadFrom(resp.Body)
		fmt.Printf("OpenAI error response: %s\n", errorBytes.String())
		return "", fmt.Errorf("OpenAI API returned status %d", resp.StatusCode)
	}

	var response struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
		Error struct {
			Message string `json:"message"`
			Type    string `json:"type"`
		} `json:"error"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		fmt.Printf("OpenAI response decode error: %v\n", err)
		return "", err
	}

	if response.Error.Message != "" {
		fmt.Printf("OpenAI API error: %s (%s)\n", response.Error.Message, response.Error.Type)
		return "", fmt.Errorf("OpenAI API error: %s", response.Error.Message)
	}

	if len(response.Choices) == 0 {
		fmt.Printf("OpenAI returned no choices\n")
		return "", fmt.Errorf("no response from OpenAI")
	}

	fmt.Printf("OpenAI API call successful\n")
	return response.Choices[0].Message.Content, nil
}

func (s *SlideService) callBedrock(prompt string) (string, error) {
	if s.config.AWSAccessKeyID == "" || s.config.AWSSecretAccessKey == "" {
		return "", fmt.Errorf("AWS credentials not configured")
	}

	// Prefer AWS SDK service if available
	if s.bedrockSDKService != nil {
		fmt.Printf("Using AWS SDK for Bedrock API call\n")
		return s.bedrockSDKService.GenerateText(prompt)
	}

	// Fallback to custom implementation
	fmt.Printf("Using custom implementation for Bedrock API call\n")
	return s.bedrockService.GenerateText(prompt)
}

// generateHTMLFromMarkdown converts markdown content to presentation-ready HTML
// using AI-powered transformation. This replaces the frontend markdown processing
// with server-side LLM-based HTML generation for better control over output.
//
// The generated HTML includes:
//   - Proper styling for presentation display
//   - Mermaid diagram placeholders with correct class names
//   - Chart.js configuration placeholders
//   - Responsive design considerations
//
// Parameters:
//   - markdown: Source markdown content to convert
//   - title: Slide title for context
//   - language: Target language for any generated text
//
// Returns:
//   - string: Generated HTML content ready for display
//   - error: Any error that occurred during generation
func (s *SlideService) generateHTMLFromMarkdown(markdown, title, language string) (string, error) {
	var prompt string
	if language == "ja" {
		prompt = fmt.Sprintf(`
以下のMarkdown形式のスライド内容を、プレゼンテーション用のHTMLに変換してください。

Markdown内容:
%s

変換要件:
1. プロフェッショナルな見た目のHTMLスライドを生成
2. Mermaidコードブロック（` + "```" + `mermaid）は <div class="mermaid">内容</div> に変換
3. Chart.js JSONコンフィグは <div class="chart-placeholder" data-chart-config='JSON'>として変換
4. レスポンシブデザインを考慮
5. 箇条書きは読みやすくスタイリング
6. 強調テキストは視覚的に目立つように
7. 完全なHTMLフラグメント（<div>で囲む）として出力

HTML:`, markdown)
	} else {
		prompt = fmt.Sprintf(`
Convert the following Markdown slide content to presentation-ready HTML.

Markdown Content:
%s

Conversion Requirements:
1. Generate professional-looking HTML slide
2. Convert Mermaid code blocks (` + "```" + `mermaid) to <div class="mermaid">content</div>
3. Convert Chart.js JSON configs to <div class="chart-placeholder" data-chart-config='JSON'>
4. Consider responsive design
5. Style bullet points for readability
6. Make emphasized text visually prominent
7. Output as complete HTML fragment (wrapped in <div>)

HTML:`, markdown)
	}

	// Use the same AI provider as for content generation
	switch s.config.AIProvider {
	case "bedrock":
		response, err := s.callBedrock(prompt)
		// Auto-fallback to OpenAI if Bedrock fails
		if err != nil {
			fmt.Printf("Bedrock HTML generation failed: %v, falling back to OpenAI\n", err)
			response, err = s.callOpenAI(prompt)
			if err != nil {
				fmt.Printf("OpenAI HTML generation fallback also failed: %v\n", err)
				return "", err
			}
			fmt.Printf("OpenAI HTML generation fallback successful\n")
		}
		return response, err
	case "openai":
		return s.callOpenAI(prompt)
	default:
		return s.callOpenAI(prompt)
	}
}