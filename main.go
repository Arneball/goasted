package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/Arneball/goasted/analyzer"
	"github.com/Arneball/goasted/formatter"
	"github.com/Arneball/goasted/rules"
)

func main() {
	var path string
	var rulesList string
	var outputFormat string

	flag.StringVar(&path, "path", ".", "Path to analyze (file or directory)")
	flag.StringVar(&rulesList, "rules", "all", "Comma-separated list of rules to run (default: all)")
	flag.StringVar(&outputFormat, "format", "text", "Output format: text or junit (default: text)")
	flag.Parse()

	// Initialize rule registry
	registry := rules.NewRegistry()
	registry.Register(rules.TestifyRule{})
	registry.Register(rules.SqlContextRule{})
	registry.Register(rules.GokitRule{})

	// Filter rules if specific rules are requested
	if rulesList != "all" && rulesList != "" {
		ruleNames := strings.Split(rulesList, ",")
		// Trim whitespace from each rule name
		for i, name := range ruleNames {
			ruleNames[i] = strings.TrimSpace(name)
		}
		registry = registry.Filter(ruleNames)
	}

	// Create analyzer
	a := analyzer.New(registry)

	// Run analysis
	violations, err := a.Analyze(path)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Error analyzing code: %v\n", err)
		os.Exit(1)
	}

	// Select formatter
	var f formatter.Formatter
	switch outputFormat {
	case "junit":
		f = formatter.JUnitFormatter{}
	case "text":
		f = formatter.TextFormatter{}
	default:
		_, _ = fmt.Fprintf(os.Stderr, "Unknown output format: %s (valid options: text, junit)\n", outputFormat)
		os.Exit(1)
	}

	// Format and output violations
	if err := f.Format(violations, os.Stdout); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Error formatting output: %v\n", err)
		os.Exit(1)
	}

	// Exit with appropriate code
	if len(violations) > 0 {
		os.Exit(1)
	}
	os.Exit(0)
}
