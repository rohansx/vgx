package scanner

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/rohansx/vgx/pkg/types"
	openai "github.com/sashabaranov/go-openai"
)

// Load .env file automatically when package is imported
func init() {
	// Try to load .env file but don't fail if missing
	if err := godotenv.Load(); err != nil {
		fmt.Println("Note: No .env file found - using environment variables")
	}
}

// AnalyzeWithOpenAI uses OpenAI's API to analyze code for security vulnerabilities
func AnalyzeWithOpenAI(fileContent string, filePath string) ([]types.Vulnerability, error) {
	// Get API key from environment variable
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("OPENAI_API_KEY environment variable not set")
	}

	// Create a client
	client := openai.NewClient(apiKey)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Create the system prompt for security analysis
	systemPrompt := `You are VibePenTester, an advanced security scanner specialized in identifying code vulnerabilities.
Your task is to analyze code for security issues including but not limited to:
- SQL injection (sqli)
- Cross-site scripting (xss)
- Command injection
- Path traversal
- Insecure deserialization
- Authorization issues
- Insecure direct object references (IDOR)
- Server-side request forgery (SSRF)
- Memory safety issues
- Race conditions

If you find a vulnerability, respond with a JSON object containing:
{
  "vulnerability_found": true,
  "issues": [
    {
      "type": "vulnerability_type",
      "description": "Detailed explanation of the issue",
      "line": line_number,
      "severity": "low|medium|high|critical",
      "recommendation": "How to fix the issue"
    }
  ]
}

If no vulnerability is found, respond with:
{
  "vulnerability_found": false
}

Be precise and focus only on clear security issues. Don't flag code quality issues unless they have security implications.`

	// Prepare the message with the code to analyze
	userMessage := fmt.Sprintf("Analyze the following code for security vulnerabilities:\n```\n%s\n```", fileContent)

	// Call the API
	resp, err := client.CreateChatCompletion(
		ctx,
		openai.ChatCompletionRequest{
			Model: "gpt-4o", // Can be configured to use different models
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleSystem,
					Content: systemPrompt,
				},
				{
					Role:    openai.ChatMessageRoleUser,
					Content: userMessage,
				},
			},
			Temperature: 0.1, // Use low temperature for more deterministic responses
		},
	)

	if err != nil {
		return nil, fmt.Errorf("OpenAI API request failed: %w", err)
	}

	// Parse the response
	aiResponse := resp.Choices[0].Message.Content
	return parseAIResponse(aiResponse, filePath)
}

// parseAIResponse converts the AI response to Vulnerability structs
func parseAIResponse(response string, filePath string) ([]types.Vulnerability, error) {
	// Simple check if vulnerability was found
	if strings.Contains(response, `"vulnerability_found": false`) {
		return nil, nil
	}

	// Extract JSON part from the response if needed
	jsonStart := strings.Index(response, "{")
	jsonEnd := strings.LastIndex(response, "}")
	if jsonStart == -1 || jsonEnd == -1 || jsonEnd <= jsonStart {
		// Fallback for malformed responses
		if strings.Contains(response, "vulnerability") && 
		   (strings.Contains(response, "found") || strings.Contains(response, "detected")) {
			return []types.Vulnerability{
				{
					File:        filePath,
					Description: "Potential security issue: " + response,
					Severity:    "medium",
					Source:      "openai",
				},
			}, nil
		}
		return nil, nil
	}

	// Simplistic parsing for demo purposes
	// In production, you would properly parse the JSON
	vulnerabilities := []types.Vulnerability{}

	// Check for issue types
	vulnTypes := []string{"sql injection", "xss", "command injection", "path traversal", 
		"insecure deserialization", "authorization", "idor", "ssrf"}
	
	for _, vType := range vulnTypes {
		if strings.Contains(strings.ToLower(response), vType) {
			// Extract line number if available
			lineNum := 0
			lineIndex := strings.Index(strings.ToLower(response), "line")
			if lineIndex != -1 {
				// Try to parse line number - simplified approach
				lineStr := response[lineIndex+4 : lineIndex+10]
				fmt.Sscanf(lineStr, "%d", &lineNum)
			}

			// Extract severity if available
			severity := "medium" // default
			for _, sev := range []string{"low", "medium", "high", "critical"} {
				if strings.Contains(strings.ToLower(response), sev) {
					severity = sev
					break
				}
			}

			// Create vulnerability entry
			vuln := types.Vulnerability{
				File:        filePath,
				Description: fmt.Sprintf("%s vulnerability detected by AI analysis", strings.Title(vType)),
				Severity:    severity,
				Line:        lineNum,
				Source:      "openai",
			}
			vulnerabilities = append(vulnerabilities, vuln)
		}
	}

	// If no specific vulnerabilities were identified but the response indicates issues
	if len(vulnerabilities) == 0 && strings.Contains(response, `"vulnerability_found": true`) {
		vulnerabilities = append(vulnerabilities, types.Vulnerability{
			File:        filePath,
			Description: "Security vulnerability detected by AI analysis",
			Severity:    "medium",
			Source:      "openai",
		})
	}

	return vulnerabilities, nil
}