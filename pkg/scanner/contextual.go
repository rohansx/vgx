package scanner

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/rohansx/vgx/pkg/types"
)

// ContextualScanRequest represents the request to the OpenAI API for contextual scanning
type ContextualScanRequest struct {
	FilePath        string   `json:"file_path"`
	FileContent     string   `json:"file_content"`
	RelatedFiles    []string `json:"related_files,omitempty"`
	CodebaseContext string   `json:"codebase_context,omitempty"`
}

// ContextualScanResponse represents the response from contextual scanning
type ContextualScanResponse struct {
	Vulnerabilities []ContextualVulnerability `json:"vulnerabilities"`
}

// ContextualVulnerability represents a vulnerability with additional context
type ContextualVulnerability struct {
	Description    string `json:"description"`
	Line           int    `json:"line"`
	Severity       string `json:"severity"`
	Recommendation string `json:"recommendation"`
	Context        string `json:"context,omitempty"`
}

// ScanWithContext performs a security scan of code with context information
func ScanWithContext(content, filePath string, contextContent []string) ([]types.Vulnerability, error) {
	// Check if OpenAI API key is available
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return runLocalScan(content, filePath)
	}

	// Check if OpenAI is disabled
	if os.Getenv("DISABLE_OPENAI") == "true" {
		return runLocalScan(content, filePath)
	}

	// Prepare context information
	contextString := ""
	if len(contextContent) > 0 {
		contextString = strings.Join(contextContent, "\n\n")
	}

	// Create request body for OpenAI
	requestData := map[string]interface{}{
		"model": "gpt-4",
		"messages": []map[string]string{
			{
				"role": "system",
				"content": `You are a security expert analyzing code for vulnerabilities. 
				Focus on identifying high-risk issues like SQL injection, XSS, command injection, 
				authentication flaws, authorization issues, path traversal, etc. 
				For each vulnerability, explain why it's a problem and provide a detailed recommendation
				for fixing it. Include the line number where the issue occurs.`,
			},
			{
				"role": "user",
				"content": fmt.Sprintf(`Analyze this code for security vulnerabilities. 
				File path: %s

				CODE:
				%s

				%s

				Respond with a JSON object with this format:
				{
					"vulnerabilities": [
						{
							"description": "Clear description of the vulnerability",
							"line": line_number,
							"severity": "low|medium|high|critical",
							"recommendation": "Detailed guidance on how to fix the issue"
						}
					]
				}

				If no vulnerabilities are found, return an empty array for "vulnerabilities".`,
					filePath, content, 
					getContextDescription(contextString)),
			},
		},
		"temperature": 0.0,
		"max_tokens": 2048,
	}

	// Convert to JSON
	jsonData, err := json.Marshal(requestData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal OpenAI request: %w", err)
	}

	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: 60 * time.Second,
	}

	// Create request
	req, err := http.NewRequest("POST", "https://api.openai.com/v1/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", apiKey))

	// Send request
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request to OpenAI: %w", err)
	}
	defer resp.Body.Close()

	// Check status code
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("OpenAI API returned non-200 status code: %d", resp.StatusCode)
	}

	// Parse response
	var openAIResponse struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&openAIResponse); err != nil {
		return nil, fmt.Errorf("failed to decode OpenAI response: %w", err)
	}

	// Process AI response
	if len(openAIResponse.Choices) == 0 {
		return nil, fmt.Errorf("OpenAI API returned empty choices")
	}

	// Parse the response content to extract vulnerabilities
	content = openAIResponse.Choices[0].Message.Content
	return parseContextualResponse(content, filePath)
}

// parseContextualResponse parses the JSON response from OpenAI
func parseContextualResponse(responseContent, filePath string) ([]types.Vulnerability, error) {
	// Extract JSON part from the response if needed
	jsonStart := strings.Index(responseContent, "{")
	jsonEnd := strings.LastIndex(responseContent, "}")
	
	if jsonStart == -1 || jsonEnd == -1 || jsonEnd <= jsonStart {
		// If no valid JSON found, try to extract basic information
		if strings.Contains(responseContent, "vulnerabilit") && 
		   strings.Contains(responseContent, "found") {
			return []types.Vulnerability{
				{
					File:        filePath,
					Description: "Potential security issue identified",
					Severity:    "medium",
					Source:      "openai",
				},
			}, nil
		}
		return nil, nil
	}

	// Extract and parse the JSON response
	jsonContent := responseContent[jsonStart:jsonEnd+1]
	
	var response ContextualScanResponse
	if err := json.Unmarshal([]byte(jsonContent), &response); err != nil {
		return nil, fmt.Errorf("failed to parse OpenAI JSON response: %w", err)
	}

	// Convert to standard vulnerabilities
	var vulnerabilities []types.Vulnerability
	for _, contextVuln := range response.Vulnerabilities {
		vulnerability := types.Vulnerability{
			File:           filePath,
			Description:    contextVuln.Description,
			Severity:       contextVuln.Severity,
			Line:           contextVuln.Line,
			Source:         "openai",
			Recommendation: contextVuln.Recommendation,
		}
		vulnerabilities = append(vulnerabilities, vulnerability)
	}

	return vulnerabilities, nil
}

// runLocalScan runs a local scan using semgrep and other available methods
func runLocalScan(content, filePath string) ([]types.Vulnerability, error) {
	// Use semgrep if available
	semgrepVulns, err := RunSemgrep(filePath)
	if err != nil {
		// Fall back to basic scanning
		return ScanContent(content, filePath)
	}
	return semgrepVulns, nil
}

// getContextDescription formats the context description
func getContextDescription(contextString string) string {
	if contextString == "" {
		return ""
	}
	return fmt.Sprintf("ADDITIONAL CONTEXT:\n%s", contextString)
} 