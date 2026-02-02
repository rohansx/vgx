package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/open-xyz/vgx/pkg/context"
	"github.com/open-xyz/vgx/pkg/detection"
	"github.com/open-xyz/vgx/pkg/scanner"
	"github.com/open-xyz/vgx/pkg/types"
)

const version = "2.0.0"

func main() {
	// Load environment variables from .env file
	godotenv.Load()

	if len(os.Args) < 2 {
		printUsage()
		os.Exit(0)
	}

	cmd := os.Args[1]
	args := os.Args[2:]

	switch cmd {
	case "scan":
		cmdScan(args)
	case "detect":
		cmdDetect(args)
	case "version", "--version", "-v":
		fmt.Printf("vgx version %s\n", version)
	case "help", "--help", "-h":
		printUsage()
	default:
		// Legacy: treat as scan with args
		cmdScan(os.Args[1:])
	}
}

func printUsage() {
	fmt.Println(`vgx - AI Code Security Scanner

USAGE:
    vgx <command> [options]

COMMANDS:
    scan      Security scan (vulnerabilities, secrets, patterns)
    detect    Detect AI-generated code
    version   Print version information
    help      Show this help message

SCAN OPTIONS:
    --changes          Scan only changed files (default: true)
    --report           Generate HTML/JSON report (default: true)
    --update-context   Update codebase context (default: true)

DETECT OPTIONS:
    --path, -p <path>  Path to scan (default: current directory)
    --format, -f       Output format: text, json (default: text)
    --threshold        AI detection threshold 0-100 (default: 70)

EXAMPLES:
    vgx scan
    vgx scan --changes=false
    vgx detect --path ./src
    vgx detect --format json
    vgx detect src/auth.ts`)
}

// cmdDetect handles the AI detection command
func cmdDetect(args []string) {
	fs := flag.NewFlagSet("detect", flag.ExitOnError)
	pathFlag := fs.String("path", "", "Path to scan")
	pFlag := fs.String("p", "", "Path to scan (short)")
	formatFlag := fs.String("format", "text", "Output format: text, json")
	fFlag := fs.String("f", "", "Output format (short)")
	thresholdFlag := fs.Float64("threshold", 70, "AI detection threshold (0-100)")
	fs.Parse(args)

	// Determine path
	path := "."
	if *pathFlag != "" {
		path = *pathFlag
	} else if *pFlag != "" {
		path = *pFlag
	} else if fs.NArg() > 0 {
		path = fs.Arg(0)
	}

	// Determine format
	format := "text"
	if *formatFlag != "text" {
		format = *formatFlag
	} else if *fFlag != "" {
		format = *fFlag
	}

	// Create detector
	detector := detection.NewDetector()
	detector.SetThreshold(*thresholdFlag / 100)

	// Check if path is file or directory
	info, err := os.Stat(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	var result detection.ScanResult
	if info.IsDir() {
		fmt.Printf("Scanning %s for AI-generated code...\n\n", path)
		result, err = detector.ScanDirectory(path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error scanning: %v\n", err)
			os.Exit(1)
		}
	} else {
		r, err := detector.AnalyzeFile(path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error analyzing: %v\n", err)
			os.Exit(1)
		}
		result = detection.ScanResult{
			FilesScanned:    1,
			AIDetected:      boolToInt(r.IsAIGenerated),
			HumanWritten:    boolToInt(!r.IsAIGenerated),
			MaxAIConfidence: r.AIConfidence,
			Results:         []detection.Result{r},
		}
	}

	// Output
	switch format {
	case "json":
		out, _ := json.MarshalIndent(result, "", "  ")
		fmt.Println(string(out))
	default:
		printDetectResults(result)
	}
}

func printDetectResults(result detection.ScanResult) {
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("  VGX AI Code Detection")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println()
	fmt.Printf("  Files scanned:     %d\n", result.FilesScanned)
	fmt.Printf("  AI-generated:      %d\n", result.AIDetected)
	fmt.Printf("  Human-written:     %d\n", result.HumanWritten)
	fmt.Printf("  AI percentage:     %.1f%%\n", result.AIPercentage)
	fmt.Printf("  Max AI confidence: %.0f%%\n", result.MaxAIConfidence*100)
	fmt.Println()

	if len(result.Results) > 0 {
		fmt.Println("  FILES")
		
		// Sort by confidence descending
		sort.Slice(result.Results, func(i, j int) bool {
			return result.Results[i].AIConfidence > result.Results[j].AIConfidence
		})
		
		for _, r := range result.Results {
			icon := "✓"
			if r.IsAIGenerated {
				icon = "🤖"
			}
			path := r.FilePath
			if len(path) > 45 {
				path = "..." + path[len(path)-42:]
			}
			fmt.Printf("     %s %-45s %5.0f%%\n", icon, path, r.AIConfidence*100)
		}
		fmt.Println()
	}

	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	if result.AIDetected > 0 {
		fmt.Printf("  🤖 %d file(s) detected as AI-generated\n", result.AIDetected)
	} else {
		fmt.Println("  ✅ No AI-generated code detected")
	}
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

// cmdScan handles the security scanning command (original functionality)
func cmdScan(args []string) {
	fs := flag.NewFlagSet("scan", flag.ExitOnError)
	changesOnly := fs.Bool("changes", true, "Scan only changed files")
	generateReport := fs.Bool("report", true, "Generate a report after scanning")
	updateContext := fs.Bool("update-context", true, "Update the codebase context after scanning")
	fs.Parse(args)

	// Initialize the context manager
	contextManager, err := context.NewContextManager()
	if err != nil {
		fmt.Printf("Error initializing context manager: %v\n", err)
		os.Exit(1)
	}

	// Clean up old reports (older than 30 days)
	if err := contextManager.CleanupOldReports(30 * 24 * time.Hour); err != nil {
		fmt.Printf("Warning: Failed to clean up old reports: %v\n", err)
	}

	// OpenAI is now optional
	hasOpenAI := os.Getenv("OPENAI_API_KEY") != ""
	if hasOpenAI {
		fmt.Println("🔍 OpenAI enhanced scanning enabled")
	} else {
		fmt.Println("📋 Running rule-based scanning (set OPENAI_API_KEY for AI-enhanced analysis)")
	}

	// Determine files to scan
	var filesToScan []string
	if fs.NArg() > 0 {
		filesToScan = fs.Args()
	} else if *changesOnly {
		files, err := contextManager.GetChangedFiles()
		if err != nil {
			fmt.Printf("Error getting changed files: %v\n", err)
			os.Exit(1)
		}
		if len(files) == 0 {
			fmt.Println("No changed files found.")
			os.Exit(0)
		}
		filesToScan = files
	} else {
		files, err := getAllFiles()
		if err != nil {
			fmt.Printf("Error getting all files: %v\n", err)
			os.Exit(1)
		}
		filesToScan = files
	}

	// Scan the files
	var allVulnerabilities []types.Vulnerability
	var scannedFiles []string

	for _, file := range filesToScan {
		info, err := os.Stat(file)
		if err != nil || info.IsDir() {
			continue
		}

		if !isTextFile(file) {
			continue
		}

		fmt.Printf("Scanning %s...\n", file)

		content, err := ioutil.ReadFile(file)
		if err != nil {
			fmt.Printf("Error: Cannot read file %s: %v\n", file, err)
			continue
		}

		contentStr := string(content)
		if !contextManager.HasFileChanged(file, contentStr) {
			fmt.Printf("Skipping unchanged file: %s\n", file)
			continue
		}

		scannedFiles = append(scannedFiles, file)

		relatedFiles := contextManager.GetRelatedFiles(file, 5)
		var contextContent []string
		for _, relatedFile := range relatedFiles {
			if fileContext, exists := contextManager.GetFileContext(relatedFile); exists {
				contextContent = append(contextContent, fmt.Sprintf("File: %s\n%s", relatedFile, fileContext.Content))
			}
		}

		vulnerabilities := scanFile(file, contentStr, hasOpenAI, contextContent)
		allVulnerabilities = append(allVulnerabilities, vulnerabilities...)

		if *updateContext {
			if err := contextManager.UpdateFileContext(file, contentStr); err != nil {
				fmt.Printf("Warning: Failed to update context for %s: %v\n", file, err)
			}
		}
	}

	// Generate report if requested
	if *generateReport && len(scannedFiles) > 0 {
		vulnerabilityMaps := make([]map[string]interface{}, 0, len(allVulnerabilities))
		for _, vuln := range allVulnerabilities {
			vulnerabilityMaps = append(vulnerabilityMaps, map[string]interface{}{
				"file":           vuln.File,
				"line":           vuln.Line,
				"description":    vuln.Description,
				"severity":       vuln.Severity,
				"source":         vuln.Source,
				"recommendation": vuln.Recommendation,
			})
		}
		if err := contextManager.GenerateReport(vulnerabilityMaps, scannedFiles); err != nil {
			fmt.Printf("Error generating report: %v\n", err)
		}
	}

	// Output the scan results
	if len(allVulnerabilities) > 0 {
		fmt.Println("\n🚨 VGX blocked commit due to vulnerabilities:")
		for _, vuln := range allVulnerabilities {
			fmt.Printf("  • [%s] %s (Source: %s)\n", vuln.File, vuln.Description, vuln.Source)
		}
		fmt.Println("\n🔧 Recommendations:")
		fmt.Println("  1. Review the flagged code")
		fmt.Println("  2. Fix the identified security issues")
		fmt.Println("  3. Commit again after resolving issues")
		os.Exit(1)
	} else if len(scannedFiles) > 0 {
		fmt.Println("\n✅ No security issues found in the scanned files!")
	}
}

func scanFile(filePath, content string, hasOpenAI bool, contextContent []string) []types.Vulnerability {
	var allVulnerabilities []types.Vulnerability

	scanner.SetSkipSemgrepErrors(true)

	if hasOpenAI {
		contextualVulns, err := scanner.ScanWithContext(content, filePath, contextContent)
		if err != nil {
			localVulns, err := scanner.ScanContent(content, filePath)
			if err == nil && len(localVulns) > 0 {
				allVulnerabilities = append(allVulnerabilities, localVulns...)
			}
		} else if len(contextualVulns) > 0 {
			allVulnerabilities = append(allVulnerabilities, contextualVulns...)
		}
	} else {
		localVulns, err := scanner.ScanContent(content, filePath)
		if err == nil && len(localVulns) > 0 {
			allVulnerabilities = append(allVulnerabilities, localVulns...)
		}
	}

	return allVulnerabilities
}

func isTextFile(filePath string) bool {
	ext := strings.ToLower(filepath.Ext(filePath))
	textExtensions := map[string]bool{
		".txt": true, ".md": true, ".js": true, ".jsx": true,
		".ts": true, ".tsx": true, ".py": true, ".java": true,
		".go": true, ".c": true, ".cpp": true, ".h": true,
		".hpp": true, ".cs": true, ".php": true, ".rb": true,
		".html": true, ".htm": true, ".css": true, ".scss": true,
		".json": true, ".xml": true, ".yaml": true, ".yml": true,
		".sh": true, ".bash": true, ".sql": true, ".rs": true,
		".kt": true, ".swift": true,
	}
	return textExtensions[ext]
}

func getAllFiles() ([]string, error) {
	cmd := exec.Command("git", "ls-files")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to list files: %v", err)
	}
	gitFiles := strings.Split(strings.TrimSpace(string(output)), "\n")
	var files []string
	for _, file := range gitFiles {
		if file != "" {
			files = append(files, file)
		}
	}
	return files, nil
}
