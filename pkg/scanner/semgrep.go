package scanner

import (
	"encoding/json"
	"os/exec"
	"strings"

	"github.com/rohansx/vgx/pkg/types"
)

// Control flags
var skipSemgrepErrors = false

// SetSkipSemgrepErrors enables or disables semgrep error reporting
func SetSkipSemgrepErrors(skip bool) {
	skipSemgrepErrors = skip
}

// SemgrepResult represents the JSON output format of semgrep
type SemgrepResult struct {
	Results struct {
		Findings []struct {
			CheckID   string `json:"check_id"`
			Path      string `json:"path"`
			Start     struct {
				Line int `json:"line"`
			} `json:"start"`
			Message  string `json:"extra.message"`
			Severity string `json:"extra.severity"`
		} `json:"findings"`
	} `json:"results"`
	Errors []struct {
		Message string `json:"message"`
	} `json:"errors"`
}

// Run Semgrep scan
func RunSemgrep(file string) ([]types.Vulnerability, error) {
	// Check if the semgrep command is available
	_, err := exec.LookPath("semgrep")
	if err != nil {
		// Semgrep is not installed
		if skipSemgrepErrors {
			return nil, nil
		}
		return []types.Vulnerability{
			{
				File:        file,
				Description: "Semgrep is not installed. Please install it with 'pip install semgrep'",
				Severity:    "info",
				Source:      "semgrep",
			},
		}, nil
	}

	// Run semgrep with JSON output for better parsing
	cmd := exec.Command("semgrep", "--json", "--config=auto", file)
	output, err := cmd.CombinedOutput()
	
	// Check for execution errors (except for finding vulnerabilities)
	if err != nil && !strings.Contains(string(output), "findings") {
		if skipSemgrepErrors {
			return nil, nil
		}
		return []types.Vulnerability{
			{
				File:        file,
				Description: "Error running semgrep: " + string(output),
				Severity:    "error",
				Source:      "semgrep",
			},
		}, nil
	}

	// Parse the JSON output
	return parseSemgrepOutput(output, file)
}

// Parse semgrep output in JSON format
func parseSemgrepOutput(output []byte, file string) ([]types.Vulnerability, error) {
	// If output is empty or doesn't look like JSON, return simple error
	if len(output) == 0 || !strings.HasPrefix(string(output), "{") {
		if len(output) > 0 && strings.TrimSpace(string(output)) != "" {
			if skipSemgrepErrors {
				return nil, nil
			}
			return []types.Vulnerability{
				{
					File:        file,
					Description: "Invalid semgrep output: " + string(output),
					Severity:    "error",
					Source:      "semgrep",
				},
			}, nil
		}
		return nil, nil
	}

	// Try to parse the full JSON structure
	var result SemgrepResult
	if err := json.Unmarshal(output, &result); err != nil {
		// If JSON parsing fails, fall back to simple detection
		if strings.Contains(string(output), "\"findings\":") && 
		   !strings.Contains(string(output), "\"findings\": []") {
			if skipSemgrepErrors {
				return nil, nil
			}
			return []types.Vulnerability{
				{
					File:        file,
					Description: "Potential security issue found by semgrep",
					Severity:    "medium",
					Source:      "semgrep",
				},
			}, nil
		}
		return nil, nil
	}

	// If there are no findings, return nil
	if len(result.Results.Findings) == 0 {
		return nil, nil
	}

	// Convert semgrep findings to our Vulnerability struct
	var vulnerabilities []types.Vulnerability
	for _, finding := range result.Results.Findings {
		vuln := types.Vulnerability{
			File:        file,
			Description: finding.Message,
			Rule:        finding.CheckID,
			Severity:    finding.Severity,
			Line:        finding.Start.Line,
			Source:      "semgrep",
		}
		
		// Default description if empty
		if vuln.Description == "" {
			vuln.Description = "Security issue detected by rule " + finding.CheckID
		}
		
		// Default severity if empty
		if vuln.Severity == "" {
			vuln.Severity = "medium"
		}
		
		vulnerabilities = append(vulnerabilities, vuln)
	}

	return vulnerabilities, nil
}