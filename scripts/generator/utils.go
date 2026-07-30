package generator

import (
	"bufio"
	"fmt"
	"go/ast"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"unicode"

	"golang.org/x/tools/go/packages"
)

func extractEntityNameFromPath(path string) string {
	base := filepath.Base(filepath.Dir(path))
	return capitalize(base)
}

func extractEntityNameForHandler(path string) string {
	filename := filepath.Base(path)                       // "department_handler.go"
	entity := strings.TrimSuffix(filename, "_handler.go") // Remove "_handler.go"
	return capitalize(entity)                             // "department"
}

func capitalize(s string) string {
	if s == "" {
		return ""
	}
	return strings.ToUpper(string(s[0])) + s[1:]
}

func toSnakeCase(str string) string {
	var result []rune
	for i, r := range str {
		if unicode.IsUpper(r) {
			if i > 0 {
				result = append(result, '_')
			}
			result = append(result, unicode.ToLower(r))
		} else {
			result = append(result, r)
		}
	}
	return string(result)
}

func formatFieldExamples(fields []StructField) string {
	var result []string
	for _, f := range fields {
		line := fmt.Sprintf("%s: %s,", f.Name, f.SampleValue)
		result = append(result, line)
	}
	return strings.Join(result, "\n\t\t")
}

func ParseAddAndUpdateStructs(sourcePath string) (*ast.StructType, *ast.StructType, error) {
	entity := extractEntityNameFromPath(sourcePath)
	addStructName := "Add" + entity
	updateStructName := "Update" + entity

	absPath, err := filepath.Abs(sourcePath)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to resolve abs path: %w", err)
	}

	cfg := &packages.Config{
		Mode:  packages.LoadAllSyntax,
		Dir:   filepath.Dir(absPath),
		Fset:  token.NewFileSet(),
		Tests: false,
	}

	pkgs, err := packages.Load(cfg, "./...")
	if err != nil || len(pkgs) == 0 {
		return nil, nil, fmt.Errorf("failed to load package: %w", err)
	}

	var addStruct, updateStruct *ast.StructType

	for _, pkg := range pkgs {
		for _, syntax := range pkg.Syntax {
			for _, decl := range syntax.Decls {
				genDecl, ok := decl.(*ast.GenDecl)
				if !ok || genDecl.Tok.String() != "type" {
					continue
				}

				for _, spec := range genDecl.Specs {
					typeSpec, ok := spec.(*ast.TypeSpec)
					if !ok {
						continue
					}

					structName := typeSpec.Name.Name
					structType, ok := typeSpec.Type.(*ast.StructType)
					if !ok {
						continue
					}

					switch structName {
					case addStructName:
						addStruct = structType
					case updateStructName:
						updateStruct = structType
					}
				}
			}
		}
	}

	if addStruct == nil || updateStruct == nil {
		fmt.Println(nil)
		return nil, nil, fmt.Errorf("could not find Add/Update struct types")
	}

	return addStruct, updateStruct, nil
}

func getGoModulePath() (string, error) {
	f, err := os.Open("go.mod")
	if err != nil {
		return "", err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "module ") {
			return strings.TrimSpace(strings.TrimPrefix(line, "module ")), nil
		}
	}

	return "", fmt.Errorf("module path not found in go.mod")
}

func inferNameField(fields []StructField) string {
	for _, f := range fields {
		if f.Name == "Name" || f.Name == "Title" || f.Name == "DepartmentName" {
			return f.Name
		}
	}
	// fallback to the first field if nothing matches
	if len(fields) > 0 {
		return fields[0].Name
	}
	return "Name"
}

// parseAddOnly parses the Go file to find and return the *ast.StructType of Add<Entity>
func parseAddOnly(sourcePath string) (*ast.StructType, error) {
	entity := extractEntityNameForHandler(sourcePath)
	addStructName := "Add" + entity

	// Assume sourcePath is like api/handler/department_handler.go → go to internal/department
	entityPackagePath := strings.Replace(sourcePath, "api/handler", "internal", 1)
	entityPackagePath = strings.Replace(entityPackagePath, filepath.Base(entityPackagePath), "", 1)

	absPath, err := filepath.Abs(entityPackagePath)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve abs path: %w", err)
	}

	cfg := &packages.Config{
		Mode:  packages.LoadAllSyntax,
		Dir:   absPath,
		Fset:  token.NewFileSet(),
		Tests: false,
	}

	pkgs, err := packages.Load(cfg, "./...")
	if err != nil || len(pkgs) == 0 {
		return nil, fmt.Errorf("failed to load package: %w", err)
	}

	for _, pkg := range pkgs {
		for _, file := range pkg.Syntax {
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
					if typeSpec.Name.Name == addStructName {
						if structType, ok := typeSpec.Type.(*ast.StructType); ok {
							return structType, nil
						}
					}
				}
			}
		}
	}

	return nil, fmt.Errorf("could not find struct %s", addStructName)
}

func buildJSONPayload(fields []StructField) string {
	var sb strings.Builder
	sb.WriteString("{")
	for i, f := range fields {
		sb.WriteString(fmt.Sprintf(`"%s": %s`, toSnakeCase(f.Name), f.SampleValue))
		if i < len(fields)-1 {
			sb.WriteString(", ")
		}
	}
	sb.WriteString("}")
	return sb.String()
}
