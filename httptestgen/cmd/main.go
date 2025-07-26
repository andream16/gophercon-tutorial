// cmd/testgen/main.go
// Test generator tool that reads JSON test cases and generates HTTP handler tests

package main

import (
	_ "embed"
	"encoding/json"
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

// TestCase represents a single test case
type (
	TestCase struct {
		CaseDescr string `json:"case_descr"`
		Request   struct {
			Method  string                 `json:"method,omitempty"`
			Path    string                 `json:"path,omitempty"`
			Body    map[string]interface{} `json:"body,omitempty"`
			Headers map[string]string      `json:"headers,omitempty"`
		} `json:"request"`
		Response struct {
			StatusCode string                 `json:"status_code"`
			Body       map[string]interface{} `json:"body,omitempty"`
			Headers    map[string]string      `json:"headers,omitempty"`
		} `json:"response"`
	}

	// EnhancedTestCase includes type information and field mappings
	EnhancedTestCase struct {
		TestCase
		RequestType    string
		ResponseType   string
		RequestFields  []FieldAssignment
		ResponseFields []FieldAssignment
	}

	// FunctionTestSpec represents all test cases for a function
	FunctionTestSpec struct {
		Func      string             `json:"func"`
		TestCases []EnhancedTestCase `json:"-"` // We'll populate this
		RawCases  []TestCase         `json:"test-cases"`
	}

	// FieldAssignment represents a Go struct field assignment
	FieldAssignment struct {
		FieldName string
		GoType    string
		Value     interface{}
		ValueCode string // The actual Go code to assign the value
	}

	// StructField represents a field in a Go struct
	StructField struct {
		Name    string
		Type    string
		JSONTag string
		GoType  string // The actual Go type (string, int, bool, etc.)
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
		ResponseTypes []string
		StructInfos   map[string]StructInfo
	}
)

var (
	inputFile     = flag.String("input", "", "Input Go file to parse")
	outputFile    = flag.String("output", "", "Output test file")
	testCasesFile = flag.String("testcases", "", "JSON file containing test cases (defaults to <input>_testcases.json)")
	requestTypes  = flag.String("request-type", "", "Comma-separated list of request types")
	responseTypes = flag.String("response-type", "", "Comma-separated list of response types")
	//go:embed test.tpl
	testTemplate string
)

func main() {
	flag.Parse()

	if *inputFile == "" || *outputFile == "" {
		log.Fatal("Both -input and -output flags are required")
	}

	// Default test cases file based on input file
	if *testCasesFile == "" {
		base := strings.TrimSuffix(*inputFile, filepath.Ext(*inputFile))
		*testCasesFile = base + "_testcases.json"
	}

	// Parse the Go file to get package name and struct information
	packageName, definedTypes, structInfos, err := parseGoFile(*inputFile)
	if err != nil {
		log.Fatalf("Error parsing Go file: %v", err)
	}

	// Parse type flags
	var reqTypes, respTypes []string
	if *requestTypes != "" {
		reqTypes = strings.Split(*requestTypes, ",")
		for i := range reqTypes {
			reqTypes[i] = strings.TrimSpace(reqTypes[i])
		}
	}
	if *responseTypes != "" {
		respTypes = strings.Split(*responseTypes, ",")
		for i := range respTypes {
			respTypes[i] = strings.TrimSpace(respTypes[i])
		}
	}

	// Validate that specified types exist in the Go file
	allTypes := append(reqTypes, respTypes...)
	for _, typeName := range allTypes {
		if !contains(definedTypes, typeName) {
			log.Printf("Warning: Type '%s' not found in %s", typeName, *inputFile)
		}
	}

	// Load test cases from JSON file
	testSpecs, err := loadTestCases(*testCasesFile)
	if err != nil {
		log.Fatalf("Error loading test cases: %v", err)
	}

	if len(testSpecs) == 0 {
		fmt.Println("No test cases found in", *testCasesFile)
		return
	}

	// Enhance test cases with type information and field mappings
	for i := range testSpecs {
		testSpecs[i].TestCases = make([]EnhancedTestCase, len(testSpecs[i].RawCases))
		for j, rawCase := range testSpecs[i].RawCases {
			requestType := inferRequestType(rawCase, reqTypes)
			responseType := inferResponseType(rawCase, respTypes)

			enhanced := EnhancedTestCase{
				TestCase:     rawCase,
				RequestType:  requestType,
				ResponseType: responseType,
			}

			// Generate field assignments for request
			if requestType != "" && len(rawCase.Request.Body) > 0 {
				if structInfo, ok := structInfos[requestType]; ok {
					enhanced.RequestFields = generateFieldAssignments(rawCase.Request.Body, structInfo)
				}
			}

			// Generate field assignments for response
			if responseType != "" && len(rawCase.Response.Body) > 0 {
				if structInfo, ok := structInfos[responseType]; ok {
					enhanced.ResponseFields = generateFieldAssignments(rawCase.Response.Body, structInfo)
				}
			}

			testSpecs[i].TestCases[j] = enhanced
		}
	}

	spec := GenerationSpec{
		PackageName:   packageName,
		FunctionSpecs: testSpecs,
		RequestTypes:  reqTypes,
		ResponseTypes: respTypes,
		StructInfos:   structInfos,
	}

	err = generateTests(spec, *outputFile)
	if err != nil {
		log.Fatalf("Error generating tests: %v", err)
	}

	fmt.Printf("Generated tests for %d function(s) in %s\n", len(testSpecs), *outputFile)
}

// parseGoFile extracts package name, defined types, and struct information from a Go file
func parseGoFile(filename string) (string, []string, map[string]StructInfo, error) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, filename, nil, parser.ParseComments)
	if err != nil {
		return "", nil, nil, err
	}

	var definedTypes []string
	structInfos := make(map[string]StructInfo)

	// Walk through declarations to find type definitions
	for _, decl := range file.Decls {
		if genDecl, ok := decl.(*ast.GenDecl); ok && genDecl.Tok == token.TYPE {
			for _, spec := range genDecl.Specs {
				if typeSpec, ok := spec.(*ast.TypeSpec); ok {
					typeName := typeSpec.Name.Name
					definedTypes = append(definedTypes, typeName)

					// If it's a struct, extract field information
					if structType, ok := typeSpec.Type.(*ast.StructType); ok {
						structInfo := parseStructType(typeName, structType)
						structInfos[typeName] = structInfo
					}
				}
			}
		}
	}

	return file.Name.Name, definedTypes, structInfos, nil
}

// parseStructType extracts field information from a struct type
func parseStructType(structName string, structType *ast.StructType) StructInfo {
	info := StructInfo{
		Name:   structName,
		Fields: []StructField{},
	}

	for _, field := range structType.Fields.List {
		for _, name := range field.Names {
			fieldInfo := StructField{
				Name:   name.Name,
				Type:   getTypeString(field.Type),
				GoType: getGoTypeString(field.Type),
			}

			// Extract JSON tag if present
			if field.Tag != nil {
				tag := field.Tag.Value
				if jsonTag := extractJSONTag(tag); jsonTag != "" {
					fieldInfo.JSONTag = jsonTag
				} else {
					// Default to lowercase field name if no json tag
					fieldInfo.JSONTag = strings.ToLower(fieldInfo.Name)
				}
			} else {
				// Default to lowercase field name if no tag
				fieldInfo.JSONTag = strings.ToLower(fieldInfo.Name)
			}

			info.Fields = append(info.Fields, fieldInfo)
		}
	}

	return info
}

// getTypeString converts an ast.Expr to a type string
func getTypeString(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.SelectorExpr:
		return getTypeString(t.X) + "." + t.Sel.Name
	case *ast.StarExpr:
		return "*" + getTypeString(t.X)
	case *ast.ArrayType:
		return "[]" + getTypeString(t.Elt)
	default:
		return "interface{}"
	}
}

// getGoTypeString returns the base Go type for code generation
func getGoTypeString(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		switch t.Name {
		case "string":
			return "string"
		case "int", "int8", "int16", "int32", "int64":
			return "int"
		case "uint", "uint8", "uint16", "uint32", "uint64":
			return "uint"
		case "float32", "float64":
			return "float"
		case "bool":
			return "bool"
		default:
			return "struct" // Custom types
		}
	case *ast.StarExpr:
		return "pointer"
	case *ast.ArrayType:
		return "slice"
	default:
		return "interface"
	}
}

// extractJSONTag extracts the json tag value from a struct tag
func extractJSONTag(tag string) string {
	// Remove backticks
	tag = strings.Trim(tag, "`")

	// Look for json:"fieldname"
	parts := strings.Split(tag, " ")
	for _, part := range parts {
		if strings.HasPrefix(part, `json:"`) {
			jsonPart := strings.TrimPrefix(part, `json:"`)
			jsonPart = strings.TrimSuffix(jsonPart, `"`)

			// Handle json:",omitempty" etc.
			if strings.Contains(jsonPart, ",") {
				jsonPart = strings.Split(jsonPart, ",")[0]
			}

			return jsonPart
		}
	}

	return ""
}

// loadTestCases loads test cases from a JSON file
func loadTestCases(filename string) ([]FunctionTestSpec, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read test cases file: %v", err)
	}

	var specs []FunctionTestSpec
	err = json.Unmarshal(data, &specs)
	if err != nil {
		return nil, fmt.Errorf("failed to parse test cases JSON: %v", err)
	}

	return specs, nil
}

// inferRequestType determines the appropriate request type for a test case
func inferRequestType(testCase TestCase, requestTypes []string) string {
	if len(testCase.Request.Body) == 0 {
		return "" // No body, no type needed
	}

	if len(requestTypes) == 0 {
		return ""
	}

	// Simple heuristic: if there's only one request type, use it
	if len(requestTypes) == 1 {
		return requestTypes[0]
	}

	// More sophisticated logic could be added here to match
	// field names in the JSON to struct fields
	return requestTypes[0] // Default to first type
}

// inferResponseType determines the appropriate response type for a test case
func inferResponseType(testCase TestCase, responseTypes []string) string {
	if len(testCase.Response.Body) == 0 {
		return "" // No body, no type needed
	}

	if len(responseTypes) == 0 {
		return ""
	}

	// Heuristic: check for error indicators
	if hasErrorFields(testCase.Response.Body) {
		// Look for ErrorResponse type
		for _, t := range responseTypes {
			if strings.Contains(strings.ToLower(t), "error") {
				return t
			}
		}
	}

	// Default to first non-error type for success cases
	for _, t := range responseTypes {
		if !strings.Contains(strings.ToLower(t), "error") {
			return t
		}
	}

	return responseTypes[0] // Fallback to first type
}

// hasErrorFields checks if the response body contains error-like fields
func hasErrorFields(body map[string]interface{}) bool {
	errorFields := []string{"error", "code", "message"}
	for field := range body {
		for _, errorField := range errorFields {
			if strings.ToLower(field) == errorField {
				return true
			}
		}
	}
	return false
}

// generateFieldAssignments creates Go code for struct field assignments
func generateFieldAssignments(jsonData map[string]interface{}, structInfo StructInfo) []FieldAssignment {
	var assignments []FieldAssignment

	for _, field := range structInfo.Fields {
		// Look for matching JSON field (by json tag or field name)
		var jsonValue interface{}
		var found bool

		// Try JSON tag first
		if jsonValue, found = jsonData[field.JSONTag]; !found {
			// Try lowercase field name
			if jsonValue, found = jsonData[strings.ToLower(field.Name)]; !found {
				// Try exact field name
				jsonValue, found = jsonData[field.Name]
			}
		}

		if found {
			assignment := FieldAssignment{
				FieldName: field.Name,
				GoType:    field.GoType,
				Value:     jsonValue,
				ValueCode: generateValueCode(jsonValue, field.GoType, field.Type),
			}
			assignments = append(assignments, assignment)
		}
	}

	return assignments
}

// generateValueCode generates Go code for assigning a value to a field
func generateValueCode(value interface{}, goType, fullType string) string {
	switch goType {
	case "string":
		if str, ok := value.(string); ok {
			return fmt.Sprintf(`"%s"`, str)
		}
		return `""`

	case "int":
		switch v := value.(type) {
		case float64:
			return fmt.Sprintf("%d", int(v))
		case int:
			return fmt.Sprintf("%d", v)
		case string:
			// Try to parse string as int
			return fmt.Sprintf(`%s`, v)
		}
		return "0"

	case "uint":
		switch v := value.(type) {
		case float64:
			return fmt.Sprintf("%d", uint(v))
		case int:
			return fmt.Sprintf("%d", uint(v))
		}
		return "0"

	case "float":
		switch v := value.(type) {
		case float64:
			return fmt.Sprintf("%f", v)
		case int:
			return fmt.Sprintf("%f", float64(v))
		}
		return "0.0"

	case "bool":
		if b, ok := value.(bool); ok {
			return fmt.Sprintf("%t", b)
		}
		return "false"

	case "struct":
		// For nested structs, we'll generate struct literal syntax
		if m, ok := value.(map[string]interface{}); ok {
			return generateStructLiteral(m, fullType)
		}
		return fmt.Sprintf("%s{}", fullType)

	case "slice":
		// For slices, generate slice literal
		if arr, ok := value.([]interface{}); ok {
			return generateSliceLiteral(arr, fullType)
		}
		return "nil"

	default:
		// Fallback to JSON marshaling for complex types
		if jsonBytes, err := json.Marshal(value); err == nil {
			return fmt.Sprintf(`json.RawMessage(%s)`, "`"+string(jsonBytes)+"`")
		}
		return "nil"
	}
}

// generateStructLiteral generates Go struct literal code
func generateStructLiteral(data map[string]interface{}, typeName string) string {
	if len(data) == 0 {
		return fmt.Sprintf("%s{}", typeName)
	}

	var fields []string
	for key, value := range data {
		// Convert key to Go field name (capitalize first letter)
		fieldName := strings.Title(key)
		valueCode := generateValueCode(value, inferGoTypeFromValue(value), "")
		fields = append(fields, fmt.Sprintf("%s: %s", fieldName, valueCode))
	}

	return fmt.Sprintf("%s{%s}", typeName, strings.Join(fields, ", "))
}

// generateSliceLiteral generates Go slice literal code
func generateSliceLiteral(data []interface{}, typeName string) string {
	if len(data) == 0 {
		return "nil"
	}

	// Extract element type from slice type (e.g., "[]string" -> "string")
	elementType := strings.TrimPrefix(typeName, "[]")

	var elements []string
	for _, value := range data {
		valueCode := generateValueCode(value, inferGoTypeFromValue(value), elementType)
		elements = append(elements, valueCode)
	}

	return fmt.Sprintf("%s{%s}", typeName, strings.Join(elements, ", "))
}

// inferGoTypeFromValue infers Go type from JSON value
func inferGoTypeFromValue(value interface{}) string {
	switch value.(type) {
	case string:
		return "string"
	case float64:
		return "float"
	case bool:
		return "bool"
	case map[string]interface{}:
		return "struct"
	case []interface{}:
		return "slice"
	default:
		return "interface"
	}
}

// contains checks if a slice contains a string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// generateTests generates the test file
func generateTests(spec GenerationSpec, outputFile string) error {
	file, err := os.Create(outputFile)
	if err != nil {
		return err
	}
	defer file.Close()

	tmpl := template.Must(template.New("test").Funcs(template.FuncMap{
		"jsonMarshal": func(v interface{}) string {
			b, _ := json.Marshal(v)
			return string(b)
		},
		"hasBody": func(body map[string]interface{}) bool {
			return len(body) > 0
		},
		"hasRequestFields": func(fields []FieldAssignment) bool {
			return len(fields) > 0
		},
		"hasResponseFields": func(fields []FieldAssignment) bool {
			return len(fields) > 0
		},
		"hasResponseType": func(responseType string) bool {
			return responseType != ""
		},
		"sanitizeName": func(s string) string {
			// Convert description to a valid test name
			s = strings.ReplaceAll(s, " ", "_")
			s = strings.ReplaceAll(s, "-", "_")
			s = strings.ReplaceAll(s, ".", "_")
			return s
		},
	}).Parse(testTemplate))

	return tmpl.Execute(file, spec)
}
