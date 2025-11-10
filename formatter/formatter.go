package formatter

import (
	"encoding/xml"
	"fmt"
	"io"

	"github.com/Arneball/goasted/rules"
)

// Formatter defines the interface for output formatters
type Formatter interface {
	Format(violations []rules.Violation, w io.Writer) error
}

// TextFormatter formats violations as plain text
type TextFormatter struct{}

func (f TextFormatter) Format(violations []rules.Violation, w io.Writer) error {
	if len(violations) == 0 {
		_, _ = fmt.Fprintln(w, "No violations found.")
		return nil
	}

	_, _ = fmt.Fprintf(w, "Found %d violation(s):\n\n", len(violations))
	for _, v := range violations {
		_, _ = fmt.Fprintf(w, "%s:%d:%d: [%s] %s\n", v.File, v.Line, v.Column, v.Rule, v.Message)
	}
	return nil
}

// JUnitFormatter formats violations as JUnit XML
type JUnitFormatter struct{}

// JUnitTestSuites JUnit XML structures
type JUnitTestSuites struct {
	XMLName xml.Name         `xml:"testsuites"`
	Suites  []JUnitTestSuite `xml:"testsuite"`
}

type JUnitTestSuite struct {
	Name     string          `xml:"name,attr"`
	Tests    int             `xml:"tests,attr"`
	Failures int             `xml:"failures,attr"`
	Errors   int             `xml:"errors,attr"`
	Time     string          `xml:"time,attr"`
	Cases    []JUnitTestCase `xml:"testcase"`
}

type JUnitTestCase struct {
	Name      string        `xml:"name,attr"`
	Classname string        `xml:"classname,attr"`
	Time      string        `xml:"time,attr"`
	Failure   *JUnitFailure `xml:"failure,omitempty"`
}

type JUnitFailure struct {
	Message string `xml:"message,attr"`
	Type    string `xml:"type,attr"`
	Content string `xml:",chardata"`
}

func (f JUnitFormatter) Format(violations []rules.Violation, w io.Writer) error {
	// Group violations by file
	fileViolations := make(map[string][]rules.Violation)
	for _, v := range violations {
		fileViolations[v.File] = append(fileViolations[v.File], v)
	}

	// Create test suites
	var suites []JUnitTestSuite

	for file, viols := range fileViolations {
		suite := JUnitTestSuite{
			Name:     file,
			Tests:    len(viols),
			Failures: len(viols),
			Errors:   0,
			Time:     "0",
		}

		for _, v := range viols {
			testCase := JUnitTestCase{
				Name:      fmt.Sprintf("%s (line %d)", v.Rule, v.Line),
				Classname: file,
				Time:      "0",
				Failure: &JUnitFailure{
					Message: v.Message,
					Type:    v.Rule,
					Content: fmt.Sprintf("%s:%d:%d: [%s] %s", v.File, v.Line, v.Column, v.Rule, v.Message),
				},
			}
			suite.Cases = append(suite.Cases, testCase)
		}

		suites = append(suites, suite)
	}

	// If no violations, create a passing test suite
	if len(violations) == 0 {
		suites = append(suites, JUnitTestSuite{
			Name:     "goasted",
			Tests:    1,
			Failures: 0,
			Errors:   0,
			Time:     "0",
			Cases: []JUnitTestCase{
				{
					Name:      "No violations",
					Classname: "goasted",
					Time:      "0",
				},
			},
		})
	}

	testSuites := JUnitTestSuites{
		Suites: suites,
	}

	output, err := xml.MarshalIndent(testSuites, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JUnit XML: %w", err)
	}

	_, _ = fmt.Fprintf(w, "%s%s\n", xml.Header, output)
	return nil
}
