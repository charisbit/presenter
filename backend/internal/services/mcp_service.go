package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"intelligent-presenter-backend/pkg/config"
)

type MCPService struct {
	config          *config.Config
	backlogWrapper  *BacklogMCPWrapper
	speechService   *SpeechService
}

func NewMCPService(cfg *config.Config) *MCPService {
	return &MCPService{
		config:         cfg,
		backlogWrapper: NewBacklogMCPWrapper(cfg),
		speechService:  NewSpeechService(cfg),
	}
}

func (s *MCPService) Start() error {
	return s.backlogWrapper.Start()
}

func (s *MCPService) Stop() error {
	return s.backlogWrapper.Stop()
}

// Backlog data retrieval methods using MCP tools

func (s *MCPService) GetProjects(backlogToken string) (interface{}, error) {
	// Use HTTP bridge to call MCP server
	return s.callBacklogToolHTTP("get_project_list", map[string]interface{}{
		"all": false,
	}, backlogToken)
}

func (s *MCPService) GetProjectOverview(projectID, backlogToken string) (interface{}, error) {
	projectData := make(map[string]interface{})
	
	// Get project details using HTTP bridge
	project, err := s.callBacklogToolHTTP("get_project", map[string]interface{}{
		"projectIdOrKey": projectID,
	}, backlogToken)
	if err != nil {
		return nil, fmt.Errorf("failed to get project: %w", err)
	}
	projectData["project"] = project
	
	// Get space info
	space, err := s.callBacklogToolHTTP("get_space", map[string]interface{}{}, backlogToken)
	if err == nil {
		projectData["space"] = space
	}
	
	// Get project users
	users, err := s.callBacklogToolHTTP("get_users", map[string]interface{}{}, backlogToken)
	if err == nil {
		projectData["users"] = users
	}
	
	return projectData, nil
}

func (s *MCPService) GetProjectProgress(projectID, backlogToken string) (interface{}, error) {
	progressData := make(map[string]interface{})
	
	// Get issues for progress analysis
	issues, err := s.callBacklogToolHTTP("get_issues", map[string]interface{}{
		"projectId": []string{projectID},
		"count":     100,
	}, backlogToken)
	if err != nil {
		return nil, fmt.Errorf("failed to get issues: %w", err)
	}
	progressData["issues"] = issues
	
	// Get issue count
	issueCount, err := s.callBacklogToolHTTP("count_issues", map[string]interface{}{
		"projectId": []string{projectID},
	}, backlogToken)
	if err == nil {
		progressData["issueCount"] = issueCount
	}
	
	return progressData, nil
}

func (s *MCPService) GetProjectIssues(projectID, backlogToken string) (interface{}, error) {
	issueData := make(map[string]interface{})
	
	// Get recent issues
	issues, err := s.callBacklogToolHTTP("get_issues", map[string]interface{}{
		"projectId": []string{projectID},
		"count":     50,
		"sort":      "updated",
		"order":     "desc",
	}, backlogToken)
	if err != nil {
		return nil, fmt.Errorf("failed to get issues: %w", err)
	}
	issueData["issues"] = issues
	
	// Get issue types
	issueTypes, err := s.callBacklogToolHTTP("get_issue_types", map[string]interface{}{
		"projectIdOrKey": projectID,
	}, backlogToken)
	if err == nil {
		issueData["issueTypes"] = issueTypes
	}
	
	// Get priorities
	priorities, err := s.callBacklogToolHTTP("get_priorities", map[string]interface{}{}, backlogToken)
	if err == nil {
		issueData["priorities"] = priorities
	}
	
	return issueData, nil
}

func (s *MCPService) GetProjectTeam(projectID, backlogToken string) (interface{}, error) {
	teamData := make(map[string]interface{})
	
	// Get project users
	users, err := s.callBacklogToolHTTP("get_users", map[string]interface{}{}, backlogToken)
	if err != nil {
		return nil, fmt.Errorf("failed to get users: %w", err)
	}
	teamData["users"] = users
	
	// Get recent activities through issues
	recentIssues, err := s.callBacklogToolHTTP("get_issues", map[string]interface{}{
		"projectId": []string{projectID},
		"count":     20,
		"sort":      "updated",
		"order":     "desc",
	}, backlogToken)
	if err == nil {
		teamData["recentActivity"] = recentIssues
	}
	
	return teamData, nil
}

func (s *MCPService) GetProjectRisks(projectID, backlogToken string) (interface{}, error) {
	riskData := make(map[string]interface{})
	
	// Get overdue/high priority issues as risks
	overdueIssues, err := s.callBacklogToolHTTP("get_issues", map[string]interface{}{
		"projectId":  []string{projectID},
		"statusId":   []string{"1", "2", "3"}, // Open statuses
		"priorityId": []string{"2", "3"},      // High/Highest priority
		"count":      30,
	}, backlogToken)
	if err != nil {
		return nil, fmt.Errorf("failed to get risk issues: %w", err)
	}
	riskData["highPriorityIssues"] = overdueIssues
	
	// Get all issues for risk analysis
	allIssues, err := s.callBacklogToolHTTP("get_issues", map[string]interface{}{
		"projectId": []string{projectID},
		"count":     100,
	}, backlogToken)
	if err == nil {
		riskData["allIssues"] = allIssues
	}
	
	return riskData, nil
}

func (s *MCPService) SynthesizeSpeech(text, language, voice string) (string, error) {
	return s.speechService.SynthesizeSpeech(text, language, voice)
}

func (s *MCPService) ServeAudioFile(filename string) (string, error) {
	return s.speechService.ServeAudioFile(filename)
}




func (s *MCPService) callBacklogToolHTTP(toolName string, arguments map[string]interface{}, accessToken ...string) (interface{}, error) {
    client := &http.Client{
        Timeout: 30 * time.Second,
    }

    // Create request for MCP HTTP Bridge
    payload := map[string]interface{}{
        "tool": toolName,
        "args": arguments,
    }
    
    // Add accessToken if provided
    if len(accessToken) > 0 && accessToken[0] != "" {
        payload["accessToken"] = accessToken[0]
    }

    jsonData, err := json.Marshal(payload)
    if err != nil {
        return nil, fmt.Errorf("failed to marshal request: %w", err)
    }

    // Use the HTTP Bridge endpoint
    url := s.config.MCPBacklogURL + "/mcp/call"
    req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
    if err != nil {
        return nil, fmt.Errorf("failed to create request: %w", err)
    }

    req.Header.Set("Content-Type", "application/json")

    resp, err := client.Do(req)
    if err != nil {
        return nil, fmt.Errorf("failed to make HTTP request: %w", err)
    }
    defer resp.Body.Close()

    bodyBytes, err := io.ReadAll(resp.Body)
    if err != nil {
        return nil, fmt.Errorf("failed to read response: %w", err)
    }

    if resp.StatusCode != http.StatusOK {
        return nil, fmt.Errorf("MCP HTTP error %d: %s", resp.StatusCode, string(bodyBytes))
    }

    // Parse bridge response { result: <jsonRaw> }
    var bridgeResp struct {
        Result json.RawMessage `json:"result"`
        Error  string          `json:"error,omitempty"`
    }
    if err := json.Unmarshal(bodyBytes, &bridgeResp); err != nil {
        return nil, fmt.Errorf("failed to unmarshal bridge response: %w", err)
    }
    if bridgeResp.Error != "" {
        return nil, fmt.Errorf("MCP bridge error: %s", bridgeResp.Error)
    }

    // Parse the actual tool result (JSON-RPC result from MCP server)
    var toolResult struct {
        Content []struct {
            Type string      `json:"type"`
            Text string      `json:"text,omitempty"`
            Data interface{} `json:"data,omitempty"`
        } `json:"content"`
    }

    if err := json.Unmarshal(bridgeResp.Result, &toolResult); err != nil {
        return nil, fmt.Errorf("failed to parse tool result: %w", err)
    }

    // Extract the actual data from the tool response
    if len(toolResult.Content) > 0 {
        if toolResult.Content[0].Data != nil {
            return toolResult.Content[0].Data, nil
        }
        if toolResult.Content[0].Text != "" {
            var data interface{}
            if err := json.Unmarshal([]byte(toolResult.Content[0].Text), &data); err == nil {
                return data, nil
            }
            return toolResult.Content[0].Text, nil
        }
    }

    return bridgeResp.Result, nil
}

func (s *MCPService) Close(ctx context.Context) error {
	return s.Stop()
}