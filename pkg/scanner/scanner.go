package scanner

import (
	"fmt"
	"os"
	"strings"

	"github.com/rohansx/vgx/pkg/cache"
	"github.com/rohansx/vgx/pkg/types"
)

// ScanFiles scans multiple files for security vulnerabilities
func ScanFiles(files []string) ([]types.Vulnerability, error) {
	var allVulnerabilities []types.Vulnerability

	for _, file := range files {
		// Try to get results from cache first
		if cachedVulns, found, err := cache.Get(file); err == nil && found {
			// Cache hit - use cached results
			fmt.Printf("Using cached results for %s\n", file)
			allVulnerabilities = append(allVulnerabilities, cachedVulns...)
			continue
		}
		
		// Cache miss - run actual scan
		vulns, err := RunSemgrep(file)
		if err != nil {
			return nil, fmt.Errorf("semgrep scan failed on %s: %w", file, err)
		}
		
		// Store results in cache for future use
		if err := cache.Store(file, vulns); err != nil {
			// Non-fatal error, just log it
			fmt.Printf("Warning: Failed to cache results for %s: %v\n", file, err)
		}
		
		allVulnerabilities = append(allVulnerabilities, vulns...)

		// Also analyze with OpenAI if configured and no vulnerabilities found with semgrep
		if shouldUseOpenAI() && len(vulns) == 0 {
			// Read file content
			content, err := os.ReadFile(file)
			if err == nil {
				openaiVulns, err := AnalyzeWithOpenAI(string(content), file)
				if err != nil {
					fmt.Printf("Warning: OpenAI analysis failed: %v\n", err)
				} else if openaiVulns != nil && len(openaiVulns) > 0 {
					allVulnerabilities = append(allVulnerabilities, openaiVulns...)
					
					// Also cache the OpenAI results
					if err := cache.Store(file, openaiVulns); err != nil {
						fmt.Printf("Warning: Failed to cache OpenAI results for %s: %v\n", file, err)
					}
				}
			}
		}
	}

	return allVulnerabilities, nil
}

// ScanContent scans code content for security vulnerabilities without requiring a file
func ScanContent(content string, identifier string) ([]types.Vulnerability, error) {
	// First attempt to use OpenAI for analysis
	var vulnerabilities []types.Vulnerability
	
	if shouldUseOpenAI() {
		openaiVulns, err := AnalyzeWithOpenAI(content, identifier)
		if err == nil && openaiVulns != nil && len(openaiVulns) > 0 {
			vulnerabilities = append(vulnerabilities, openaiVulns...)
		}
	}
	
	// More scanners can be added here
	
	return vulnerabilities, nil
}

// shouldUseOpenAI checks if OpenAI integration should be used
func shouldUseOpenAI() bool {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return false
	}
	
	// Check for explicit disabling
	disableOpenAI := os.Getenv("DISABLE_OPENAI")
	if strings.ToLower(disableOpenAI) == "true" || disableOpenAI == "1" {
		return false
	}
	
	return true
}