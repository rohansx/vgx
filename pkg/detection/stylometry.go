package detection

import (
	"math"
	"regexp"
	"strings"
)

// StyleMetrics contains extracted style features from code
type StyleMetrics struct {
	NamingConsistency      float64 // High = AI signal
	IndentationConsistency float64 // High = AI signal
	CommentDensity         float64 // ~15% = AI-like
	AvgLineLength          float64 // ~45 chars = AI-like
	LineLengthVariance     float64 // Low = AI signal
	BoilerplateRatio       float64 // High = AI signal
	EmptyLineRatio         float64 // Consistent = AI signal
}

// StyleAnalyzer analyzes code for AI generation patterns
type StyleAnalyzer struct {
	boilerplatePatterns []*regexp.Regexp
}

// NewStyleAnalyzer creates a new analyzer
func NewStyleAnalyzer() *StyleAnalyzer {
	patterns := []string{
		`try\s*\{[\s\S]*?catch`,
		`if\s*\(\s*!\s*\w+\s*\)\s*\{?\s*return`,
		`async\s+function\s+\w+\s*\([^)]*\)\s*\{`,
		`const\s+\w+\s*=\s*async\s*\([^)]*\)\s*=>`,
		`export\s+(default\s+)?(function|class|const)`,
		`import\s*\{[^}]+\}\s*from`,
		`if\s+err\s*!=\s*nil\s*\{`,           // Go error handling
		`defer\s+\w+\.(Close|Unlock|Done)\(`, // Go defer
		`func\s+\(\w+\s+\*?\w+\)\s+\w+\(`,    // Go methods
	}

	compiled := make([]*regexp.Regexp, 0, len(patterns))
	for _, p := range patterns {
		if re, err := regexp.Compile(p); err == nil {
			compiled = append(compiled, re)
		}
	}

	return &StyleAnalyzer{boilerplatePatterns: compiled}
}

// Analyze extracts style metrics from code
func (a *StyleAnalyzer) Analyze(code string) StyleMetrics {
	lines := strings.Split(code, "\n")
	nonEmptyLines := make([]string, 0)
	for _, l := range lines {
		if strings.TrimSpace(l) != "" {
			nonEmptyLines = append(nonEmptyLines, l)
		}
	}

	return StyleMetrics{
		NamingConsistency:      a.analyzeNaming(code),
		IndentationConsistency: a.analyzeIndentation(lines),
		CommentDensity:         a.analyzeComments(lines),
		AvgLineLength:          a.avgLineLength(nonEmptyLines),
		LineLengthVariance:     a.lineLengthVariance(nonEmptyLines),
		BoilerplateRatio:       a.analyzeBoilerplate(code, len(lines)),
		EmptyLineRatio:         a.emptyLineRatio(lines),
	}
}

func (a *StyleAnalyzer) analyzeNaming(code string) float64 {
	// Extract identifiers
	re := regexp.MustCompile(`\b([a-z][a-zA-Z0-9_]*)\b`)
	matches := re.FindAllString(code, -1)

	if len(matches) < 5 {
		return 0.5
	}

	camelCase := 0
	snakeCase := 0
	camelRe := regexp.MustCompile(`^[a-z]+([A-Z][a-z]*)*$`)
	
	for _, m := range matches {
		if camelRe.MatchString(m) {
			camelCase++
		}
		if strings.Contains(m, "_") && strings.ToLower(m) == m {
			snakeCase++
		}
	}

	total := float64(len(matches))
	dominant := float64(max(camelCase, snakeCase))
	return dominant / total
}

func (a *StyleAnalyzer) analyzeIndentation(lines []string) float64 {
	indents := make([]int, 0)
	for _, line := range lines {
		if strings.TrimSpace(line) != "" {
			indent := len(line) - len(strings.TrimLeft(line, " \t"))
			indents = append(indents, indent)
		}
	}

	if len(indents) == 0 {
		return 0.5
	}

	// Check for consistent indent unit (2 or 4 spaces)
	indentUnit := 4
	for _, i := range indents {
		if i%2 == 0 && i%4 != 0 {
			indentUnit = 2
			break
		}
	}

	consistent := 0
	for _, i := range indents {
		if i%indentUnit == 0 {
			consistent++
		}
	}

	return float64(consistent) / float64(len(indents))
}

func (a *StyleAnalyzer) analyzeComments(lines []string) float64 {
	commentLines := 0
	for _, l := range lines {
		trimmed := strings.TrimSpace(l)
		if strings.HasPrefix(trimmed, "//") || 
		   strings.HasPrefix(trimmed, "#") || 
		   strings.HasPrefix(trimmed, "/*") ||
		   strings.HasPrefix(trimmed, "*") {
			commentLines++
		}
	}
	if len(lines) == 0 {
		return 0
	}
	return float64(commentLines) / float64(len(lines))
}

func (a *StyleAnalyzer) avgLineLength(lines []string) float64 {
	if len(lines) == 0 {
		return 0
	}
	total := 0
	for _, l := range lines {
		total += len(l)
	}
	return float64(total) / float64(len(lines))
}

func (a *StyleAnalyzer) lineLengthVariance(lines []string) float64 {
	if len(lines) < 2 {
		return 0
	}
	
	avg := a.avgLineLength(lines)
	sumSquares := 0.0
	for _, l := range lines {
		diff := float64(len(l)) - avg
		sumSquares += diff * diff
	}
	return math.Sqrt(sumSquares / float64(len(lines)))
}

func (a *StyleAnalyzer) analyzeBoilerplate(code string, lineCount int) float64 {
	matches := 0
	for _, re := range a.boilerplatePatterns {
		matches += len(re.FindAllString(code, -1))
	}
	if lineCount == 0 {
		return 0
	}
	ratio := float64(matches) / (float64(lineCount) / 10)
	return math.Min(ratio, 1.0)
}

func (a *StyleAnalyzer) emptyLineRatio(lines []string) float64 {
	empty := 0
	for _, l := range lines {
		if strings.TrimSpace(l) == "" {
			empty++
		}
	}
	if len(lines) == 0 {
		return 0
	}
	return float64(empty) / float64(len(lines))
}

// AIConfidence calculates the probability that code is AI-generated
func (m StyleMetrics) AIConfidence() float64 {
	weights := map[string]float64{
		"naming":      0.20,
		"indentation": 0.20,
		"boilerplate": 0.20,
		"comment":     0.10,
		"linevar":     0.15,
		"emptyline":   0.10,
		"linelen":     0.05,
	}

	signals := map[string]float64{
		"naming":      m.NamingConsistency,
		"indentation": m.IndentationConsistency,
		"boilerplate": m.BoilerplateRatio,
		"comment":     1 - math.Abs(m.CommentDensity-0.15)*5,
		"linevar":     math.Max(0, 1-m.LineLengthVariance/30),
		"emptyline":   1 - math.Abs(m.EmptyLineRatio-0.15)*5,
		"linelen":     1 - math.Abs(m.AvgLineLength-45)/45,
	}

	// Clip signals to [0, 1]
	for k, v := range signals {
		signals[k] = math.Max(0, math.Min(1, v))
	}

	probability := 0.0
	for k, w := range weights {
		probability += signals[k] * w
	}

	return probability
}

// ConfidenceLevel returns a human-readable confidence level
func (m StyleMetrics) ConfidenceLevel() string {
	conf := m.AIConfidence()
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

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
