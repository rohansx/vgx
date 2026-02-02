package detection

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Result contains the analysis result for a file
type Result struct {
	FilePath        string         `json:"file_path"`
	AIConfidence    float64        `json:"ai_confidence"`
	ConfidenceLevel string         `json:"confidence_level"`
	StyleScore      float64        `json:"style_score"`
	PatternScore    float64        `json:"pattern_score"`
	Patterns        []PatternMatch `json:"patterns,omitempty"`
	IsAIGenerated   bool           `json:"is_ai_generated"`
	LinesOfCode     int            `json:"lines_of_code"`
}

// ScanResult contains results for multiple files
type ScanResult struct {
	FilesScanned    int       `json:"files_scanned"`
	AIDetected      int       `json:"ai_detected"`
	HumanWritten    int       `json:"human_written"`
	MaxAIConfidence float64   `json:"max_ai_confidence"`
	AIPercentage    float64   `json:"ai_percentage"`
	Results         []Result  `json:"results"`
}

// Detector combines multiple detection methods
type Detector struct {
	styleAnalyzer   *StyleAnalyzer
	patternDetector *PatternDetector
	threshold       float64
}

// NewDetector creates a new combined detector
func NewDetector() *Detector {
	return &Detector{
		styleAnalyzer:   NewStyleAnalyzer(),
		patternDetector: NewPatternDetector(),
		threshold:       0.70, // 70% = AI-generated
	}
}

// SetThreshold sets the AI detection threshold
func (d *Detector) SetThreshold(t float64) {
	d.threshold = t
}

// AnalyzeCode analyzes a string of code
func (d *Detector) AnalyzeCode(code string, filename string) Result {
	lines := len(strings.Split(code, "\n"))
	
	// Style analysis
	metrics := d.styleAnalyzer.Analyze(code)
	styleScore := metrics.AIConfidence()
	
	// Pattern detection
	patterns, patternScore := d.patternDetector.Detect(code)
	
	// Combined score (weighted average)
	combined := (styleScore*0.45 + patternScore*0.55)
	
	return Result{
		FilePath:        filename,
		AIConfidence:    combined,
		ConfidenceLevel: d.confidenceLevel(combined),
		StyleScore:      styleScore,
		PatternScore:    patternScore,
		Patterns:        patterns,
		IsAIGenerated:   combined >= d.threshold,
		LinesOfCode:     lines,
	}
}

// AnalyzeFile analyzes a single file
func (d *Detector) AnalyzeFile(filepath string) (Result, error) {
	content, err := os.ReadFile(filepath)
	if err != nil {
		return Result{}, err
	}
	return d.AnalyzeCode(string(content), filepath), nil
}

// ScanDirectory scans a directory for AI-generated code
func (d *Detector) ScanDirectory(root string) (ScanResult, error) {
	result := ScanResult{
		Results: make([]Result, 0),
	}
	
	extensions := map[string]bool{
		".go": true, ".py": true, ".js": true, ".ts": true, ".tsx": true,
		".jsx": true, ".java": true, ".kt": true, ".rs": true, ".rb": true,
		".php": true, ".swift": true, ".cs": true, ".cpp": true, ".c": true,
	}
	
	skipDirs := map[string]bool{
		"node_modules": true, ".git": true, "vendor": true, "dist": true,
		"build": true, "__pycache__": true, ".next": true, "target": true,
	}
	
	totalAILines := 0
	totalLines := 0
	
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		
		if info.IsDir() {
			if skipDirs[info.Name()] {
				return filepath.SkipDir
			}
			return nil
		}
		
		ext := filepath.Ext(path)
		if !extensions[ext] {
			return nil
		}
		
		r, err := d.AnalyzeFile(path)
		if err != nil {
			return nil
		}
		
		result.Results = append(result.Results, r)
		result.FilesScanned++
		
		if r.IsAIGenerated {
			result.AIDetected++
			totalAILines += r.LinesOfCode
		} else {
			result.HumanWritten++
		}
		totalLines += r.LinesOfCode
		
		if r.AIConfidence > result.MaxAIConfidence {
			result.MaxAIConfidence = r.AIConfidence
		}
		
		return nil
	})
	
	if totalLines > 0 {
		result.AIPercentage = float64(totalAILines) / float64(totalLines) * 100
	}
	
	return result, err
}

func (d *Detector) confidenceLevel(conf float64) string {
	switch {
	case conf > 0.85:
		return "very_high"
	case conf > 0.70:
		return "high"
	case conf > 0.50:
		return "medium"
	case conf > 0.30:
		return "low"
	default:
		return "very_low"
	}
}

// PrintResult prints a single result
func (r Result) Print() {
	icon := "✓"
	if r.IsAIGenerated {
		icon = "🤖"
	}
	fmt.Printf("%s %-50s %5.0f%%\n", icon, truncate(r.FilePath, 50), r.AIConfidence*100)
}

// PrintSummary prints scan summary
func (s ScanResult) PrintSummary() {
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("  VGX AI Code Detection")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println()
	fmt.Printf("  Files scanned:     %d\n", s.FilesScanned)
	fmt.Printf("  AI-generated:      %d\n", s.AIDetected)
	fmt.Printf("  Human-written:     %d\n", s.HumanWritten)
	fmt.Printf("  AI percentage:     %.1f%%\n", s.AIPercentage)
	fmt.Printf("  Max AI confidence: %.0f%%\n", s.MaxAIConfidence*100)
	fmt.Println()
	
	fmt.Println("  FILES")
	for _, r := range s.Results {
		r.Print()
	}
	fmt.Println()
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return "..." + s[len(s)-max+3:]
}
