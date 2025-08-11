// cmd/testgen/main.go
// Test generator tool that reads JSON test cases and generates HTTP handler tests

package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
)

// TestCase represents a single test case
type (
	TestCase struct {
		CaseDescr string   `json:"case_descr"`
		Request   Request  `json:"request"`
		Response  Response `json:"response"`
	}

	Request struct {
		Method string         `json:"method,omitempty"`
		Path   string         `json:"path,omitempty"`
		Body   map[string]any `json:"body,omitempty"`
	}

	Response struct {
		StatusCode string         `json:"status_code"`
		Body       map[string]any `json:"body,omitempty"`
	}

	// EnhancedTestCase includes type information and field mappings
	EnhancedTestCase struct {
		TestCase
		RequestType    string
		RequestFields  []FieldAssignment
		ResponseFields []FieldAssignment
	}

	// FunctionTestSpec represents all test cases for a function
	FunctionTestSpec struct {
		Func      string             `json:"func"`
		TestCases []EnhancedTestCase `json:"-"`
		RawCases  []TestCase         `json:"test-cases"`
	}

	// FieldAssignment represents a Go struct field assignment
	FieldAssignment struct {
		FieldName string
		GoType    string
		Value     any
		ValueCode string
	}

	// StructField represents a field in a Go struct
	StructField struct {
		Name    string
		Type    string
		JSONTag string
		GoType  string
	}

	// StructInfo contains information about a Go struct
	StructInfo struct {
		Name   string
		Fields []StructField
	}

	// GenerationSpec holds all the information needed for test generation
	GenerationSpec struct {
		PackageName   string
		FunctionSpecs []FunctionTestSpec
		RequestTypes  []string
		StructInfos   map[string]StructInfo
	}
)

func main() {
	if err := Main(); err != nil {
		log.Fatalf("could not run: %v", err)
	}
}

func Main() error {
	cfg, err := initConfig()
	if err != nil {
		return fmt.Errorf("could not init config: %w", err)
	}

	// Parse the Go file to get package name and struct information
	packageName, definedTypes, structInfos, err := parseGoFile(cfg.inputFile)
	if err != nil {
		return fmt.Errorf("could not parse input file %s: %w", cfg.inputFile, err)
	}

	var dtm = make(map[string]struct{}, len(definedTypes))
	for _, dt := range definedTypes {
		dtm[dt] = struct{}{}
	}

	// Check if passed request types are supported.
	for _, t := range cfg.requestTypes {
		if _, ok := dtm[t]; !ok {
			return fmt.Errorf("unsupported invalid request type: %s", t)
		}
	}

	// Load test cases from JSON file.
	testSpecs, err := loadTestCases(cfg.testCasesFile)
	if err != nil {
		return fmt.Errorf("could not load test cases: %w", err)
	}

	if len(testSpecs) == 0 {
		return fmt.Errorf("no test cases found in %s", cfg.testCasesFile)
	}

	// Prepare tests meta.
	spec := prepareSpecs(packageName, testSpecs, cfg.requestTypes, structInfos)

	// Generate.
	if err = generateTests(spec, cfg.outputFile); err != nil {
		return fmt.Errorf("could not generate test cases: %w", err)
	}

	fmt.Printf("Generated tests for %d function(s) in %s\n", len(testSpecs), cfg.outputFile)
	return nil
}

// loadTestCases loads test cases from a JSON file
func loadTestCases(filename string) ([]FunctionTestSpec, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read test cases file: %w", err)
	}

	var specs []FunctionTestSpec
	if err = json.Unmarshal(data, &specs); err != nil {
		return nil, fmt.Errorf("failed to parse test cases JSON: %w", err)
	}

	return specs, nil
}

// inferRequestType determines the appropriate request type for a test case
func inferRequestType(testCase TestCase, requestTypes []string) string {
	if len(testCase.Request.Body) == 0 || len(requestTypes) == 0 {
		return ""
	}

	// If there's only one request type, use it
	if len(requestTypes) == 1 {
		return requestTypes[0]
	}

	// More sophisticated logic could be added here to match
	// field names in the JSON to struct fields
	return requestTypes[0]
}
