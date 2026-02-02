package detection

import (
	"regexp"
	"strings"
)

// PatternMatch represents a detected AI pattern
type PatternMatch struct {
	Name       string
	Confidence float64
	LineStart  int
	LineEnd    int
	Snippet    string
}

// PatternDetector detects known AI code patterns
type PatternDetector struct {
	patterns []patternDef
}

type patternDef struct {
	name    string
	pattern *regexp.Regexp
	weight  float64
}

// NewPatternDetector creates a new pattern detector
func NewPatternDetector() *PatternDetector {
	defs := []struct {
		name    string
		pattern string
		weight  float64
	}{
		// Error handling patterns
		{"copilot_try_catch", `try\s*\{[^}]+\}\s*catch\s*\(\s*(?:error|err|e)\s*(?::\s*\w+)?\s*\)\s*\{[^}]*(?:console\.(?:error|log)|throw)[^}]*\}`, 0.15},
		{"standard_error_throw", `throw\s+new\s+Error\s*\(\s*['"\x60](?:Failed to|Unable to|Error|Invalid|Cannot)[^'"\x60]+['"\x60]\s*\)`, 0.12},
		
		// Async patterns
		{"async_await_fetch", `async\s+(?:function\s+)?\w+\s*\([^)]*\)\s*(?::\s*Promise<[^>]+>)?\s*\{[^}]*await\s+fetch`, 0.10},
		{"promise_chain", `\.then\s*\(\s*(?:\([^)]*\)|[a-z]+)\s*=>\s*\{?[^}]*\}\s*\)\s*\.catch`, 0.08},
		
		// Comment patterns
		{"jsdoc_complete", `/\*\*\s*\n(?:\s*\*\s*@\w+[^\n]*\n)+\s*\*/`, 0.10},
		{"inline_explanation", `//\s*[A-Z][a-z]+(?:\s+[a-z]+){3,}`, 0.08},
		
		// Function patterns
		{"arrow_with_types", `const\s+\w+\s*=\s*(?:async\s*)?\([^)]*:\s*\w+[^)]*\)\s*(?::\s*\w+(?:<[^>]+>)?)?\s*=>`, 0.10},
		{"export_default_function", `export\s+default\s+(?:async\s+)?function\s+\w+`, 0.06},
		
		// React patterns
		{"use_effect_deps", `useEffect\s*\(\s*\(\s*\)\s*=>\s*\{[^}]+\}\s*,\s*\[[^\]]*\]\s*\)`, 0.08},
		{"use_state_destructure", `const\s*\[\s*\w+\s*,\s*set[A-Z]\w+\s*\]\s*=\s*useState`, 0.08},
		
		// Go patterns
		{"go_error_check", `if\s+err\s*!=\s*nil\s*\{[^}]*return[^}]*\}`, 0.12},
		{"go_defer", `defer\s+(?:\w+\.)?(?:Close|Unlock|Done)\s*\(\s*\)`, 0.08},
		{"go_struct_init", `\w+\s*:=\s*&?\w+\{\s*\n(?:\s*\w+:\s*[^,]+,?\s*\n)+\s*\}`, 0.08},
		
		// Python patterns  
		{"python_docstring", `"""[^"]+(?:Args:|Returns:|Raises:)[^"]+"""`, 0.10},
		{"python_type_hints", `def\s+\w+\s*\([^)]*:\s*\w+[^)]*\)\s*->\s*\w+:`, 0.08},
		
		// Generic AI signatures
		{"numbered_steps", `//\s*(?:Step\s+)?\d+[.:]\s*[A-Z]`, 0.06},
		{"todo_ai_style", `//\s*TODO:\s*[A-Z][a-z]+\s+[a-z]+`, 0.05},
	}

	pd := &PatternDetector{patterns: make([]patternDef, 0)}
	for _, d := range defs {
		if re, err := regexp.Compile(d.pattern); err == nil {
			pd.patterns = append(pd.patterns, patternDef{
				name:    d.name,
				pattern: re,
				weight:  d.weight,
			})
		}
	}
	return pd
}

// Detect finds AI patterns in code
func (pd *PatternDetector) Detect(code string) ([]PatternMatch, float64) {
	var matches []PatternMatch
	totalWeight := 0.0

	for _, p := range pd.patterns {
		for _, m := range p.pattern.FindAllStringIndex(code, -1) {
			start := strings.Count(code[:m[0]], "\n") + 1
			end := strings.Count(code[:m[1]], "\n") + 1
			
			snippet := code[m[0]:m[1]]
			if len(snippet) > 80 {
				snippet = snippet[:80] + "..."
			}

			matches = append(matches, PatternMatch{
				Name:       p.name,
				Confidence: p.weight,
				LineStart:  start,
				LineEnd:    end,
				Snippet:    snippet,
			})
			totalWeight += p.weight
		}
	}

	// Cap score at 1.0
	score := totalWeight
	if score > 1.0 {
		score = 1.0
	}

	return matches, score
}
