package main

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"text/template"
)

var (
	//go:embed test.tpl
	testTemplate string
)

// generateValueCode generates Go code for assigning a value to a field
func generateValueCode(value any, goType, fullType string) string {
	switch goType {
	case "string":
		str, ok := value.(string)
		if !ok {
			return `""`
		}
		return fmt.Sprintf(`"%s"`, str)
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
		if m, ok := value.(map[string]any); ok {
			return generateStructLiteral(m, fullType)
		}
		return fmt.Sprintf("%s{}", fullType)

	case "slice":
		// For slices, generate slice literal
		if arr, ok := value.([]any); ok {
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
func generateStructLiteral(data map[string]any, typeName string) string {
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
func generateSliceLiteral(data []any, typeName string) string {
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
func inferGoTypeFromValue(value any) string {
	switch value.(type) {
	case string:
		return "string"
	case float64:
		return "float"
	case bool:
		return "bool"
	case map[string]any:
		return "struct"
	case []any:
		return "slice"
	default:
		return "interface"
	}
}

// generateTests generates the test file
func generateTests(spec GenerationSpec, outputFile string) error {
	file, err := os.Create(outputFile)
	if err != nil {
		return err
	}
	defer file.Close()

	tmpl := template.Must(template.New("test").Funcs(template.FuncMap{
		"jsonMarshal": func(v any) string {
			b, _ := json.Marshal(v)
			return string(b)
		},
		"hasBody": func(body map[string]any) bool {
			return len(body) > 0
		},
		"hasRequestFields": func(fields []FieldAssignment) bool {
			return len(fields) > 0
		},
		"hasResponseFields": func(fields []FieldAssignment) bool {
			return len(fields) > 0
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

func generateFieldAssignments(jsonData map[string]any, structInfo StructInfo) []FieldAssignment {
	var assignments []FieldAssignment

	for _, field := range structInfo.Fields {
		// Look for matching JSON field (by json tag or field name)
		var (
			jsonValue any
			found     bool
		)

		// Try JSON tag first
		if jsonValue, found = jsonData[field.JSONTag]; !found {
			// Try lowercase field name
			if jsonValue, found = jsonData[strings.ToLower(field.Name)]; !found {
				// Try exact field name
				jsonValue, found = jsonData[field.Name]
			}
		}

		if !found {
			continue
		}

		assignments = append(assignments, FieldAssignment{
			FieldName: field.Name,
			GoType:    field.GoType,
			Value:     jsonValue,
			ValueCode: generateValueCode(jsonValue, field.GoType, field.Type),
		})
	}

	return assignments
}

func prepareSpecs(
	pkgName string,
	testSpecs []FunctionTestSpec,
	reqTypes []string,
	structInfos map[string]StructInfo,
) GenerationSpec {
	// Enhance test cases with type information and field mappings
	for i := range testSpecs {
		testSpecs[i].TestCases = make([]EnhancedTestCase, len(testSpecs[i].RawCases))
		for j, rawCase := range testSpecs[i].RawCases {
			requestType := inferRequestType(rawCase, reqTypes)

			enhanced := EnhancedTestCase{
				TestCase:    rawCase,
				RequestType: requestType,
			}

			// Generate field assignments for request
			if requestType != "" && len(rawCase.Request.Body) > 0 {
				if structInfo, ok := structInfos[requestType]; ok {
					enhanced.RequestFields = generateFieldAssignments(rawCase.Request.Body, structInfo)
				}
			}

			testSpecs[i].TestCases[j] = enhanced
		}
	}

	return GenerationSpec{
		PackageName:   pkgName,
		FunctionSpecs: testSpecs,
		RequestTypes:  reqTypes,
		StructInfos:   structInfos,
	}
}
