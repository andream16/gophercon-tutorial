package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"strings"
)

// parseGoFile extracts package name, defined types, and struct information from a Go file
func parseGoFile(filename string) (string, []string, map[string]StructInfo, error) {
	file, err := parser.ParseFile(token.NewFileSet(), filename, nil, parser.ParseComments)
	if err != nil {
		return "", nil, nil, fmt.Errorf("could not parse file %s: %w", filename, err)
	}

	var (
		packageName  = file.Name.Name
		definedTypes []string
		structInfos  = make(map[string]StructInfo)
	)

	// Walk through declarations to find type definitions
	for _, decl := range file.Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok || genDecl.Tok != token.TYPE {
			continue
		}

		for _, spec := range genDecl.Specs {
			typeSpec, ok := spec.(*ast.TypeSpec)
			if !ok {
				continue
			}

			typeName := typeSpec.Name.Name
			definedTypes = append(definedTypes, typeName)

			// If it's a struct, extract field information
			structType, ok := typeSpec.Type.(*ast.StructType)
			if !ok {
				continue
			}

			structInfos[typeName] = parseStructType(typeName, structType)
		}
	}

	return packageName, definedTypes, structInfos, nil
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

			// Default to lowercase field name if no tag
			fieldInfo.JSONTag = strings.ToLower(fieldInfo.Name)

			// Extract JSON tag if present
			if field.Tag != nil {
				if jsonTag := extractJSONTag(field.Tag.Value); jsonTag != "" {
					fieldInfo.JSONTag = jsonTag
				}
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
		return "any"
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
	// Look for json:"fieldname"
	parts := strings.Split(strings.Trim(tag, "`"), " ")
	for _, part := range parts {
		if !strings.HasPrefix(part, `json:"`) {
			continue
		}

		jsonPart := strings.TrimPrefix(part, `json:"`)
		jsonPart = strings.TrimSuffix(jsonPart, `"`)

		// Handle json:",omitempty" etc.
		if strings.Contains(jsonPart, ",") {
			jsonPart = strings.Split(jsonPart, ",")[0]
		}

		return jsonPart
	}

	return ""
}
