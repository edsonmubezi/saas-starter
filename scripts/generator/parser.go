package generator

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"golang.org/x/tools/go/packages"
)

// ExtractUsecaseInterfaceName returns the first interface ending in "UseCase".
func ExtractUsecaseInterfaceName(filePath string) string {
	cfg := &packages.Config{
		Mode: packages.NeedTypes | packages.NeedSyntax | packages.NeedTypesInfo,
		Dir:  filepath.Dir(filePath),
		Fset: token.NewFileSet(),
	}

	pkgs, _ := packages.Load(cfg, "./...")
	for _, pkg := range pkgs {
		for _, file := range pkg.Syntax {
			for _, decl := range file.Decls {
				if genDecl, ok := decl.(*ast.GenDecl); ok && genDecl.Tok.String() == "type" {
					for _, spec := range genDecl.Specs {
						if ts, ok := spec.(*ast.TypeSpec); ok {
							if strings.HasSuffix(ts.Name.Name, "UseCase") {
								return ts.Name.Name
							}
						}
					}
				}
			}
		}
	}
	return ""
}

func FindStructByPrefixOrNil(dirPath string, prefix string) *ast.StructType {
	structs, _ := ParseAllStructs(dirPath) // <-- parse all Go files
	for name, node := range structs {
		if strings.HasPrefix(name, prefix) {
			return node
		}
	}
	return nil
}

// StructFields extracts fields from an AST struct
func StructFields(structType *ast.StructType) []StructField {
	var fields []StructField
	for _, f := range structType.Fields.List {
		if len(f.Names) == 0 {
			continue
		}
		name := f.Names[0].Name
		typeStr := exprToString(f.Type)
		fields = append(fields, StructField{
			Name:        name,
			Type:        typeStr,
			SampleValue: sampleValueForType(typeStr),
		})
	}
	return fields
}

// FirstField returns the first visible field from the struct
func FirstField(structType *ast.StructType) StructField {
	fields := StructFields(structType)
	if len(fields) > 0 {
		return fields[0]
	}
	return StructField{Name: "Unknown", Type: "string", SampleValue: `"example"`}
}

// Converts AST expression to string
func exprToString(expr ast.Expr) string {
	switch e := expr.(type) {
	case *ast.Ident:
		return e.Name
	case *ast.SelectorExpr:
		return exprToString(e.X) + "." + e.Sel.Name
	case *ast.StarExpr:
		return "*" + exprToString(e.X)
	case *ast.ArrayType:
		return "[]" + exprToString(e.Elt)
	default:
		return "interface{}"
	}
}

// Guess sample values
func sampleValueForType(t string) string {
	switch t {
	case "string":
		return `"example"`
	case "int", "int64":
		return "1"
	case "bool":
		return "true"
	case "time.Time":
		return "time.Now()"
	case "sql.NullTime":
		return "sql.NullTime{Valid: false}"
	default:
		if strings.HasPrefix(t, "*") {
			return "nil"
		}
		return "nil"
	}
}

// WriteTemplateToFile renders a template with data and writes it to the target file path.
func WriteTemplateToFile(tmpl *template.Template, data any, filePath string) error {
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	return tmpl.Execute(file, data)
}

func ParseInterfaceMethods(sourcePath string) ([]string, error) {
	absPath, err := filepath.Abs(sourcePath)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path: %w", err)
	}

	cfg := &packages.Config{
		Mode:  packages.LoadAllSyntax,
		Dir:   filepath.Dir(absPath),
		Fset:  token.NewFileSet(),
		Tests: false,
	}

	pkgs, err := packages.Load(cfg, "./...")
	if err != nil || len(pkgs) == 0 {
		return nil, fmt.Errorf("failed to load package: %w", err)
	}

	var methodNames []string

	for _, pkg := range pkgs {
		scope := pkg.Types.Scope()
		for _, name := range scope.Names() {
			obj := scope.Lookup(name)
			if obj == nil {
				continue
			}

			if iface, ok := obj.Type().Underlying().(*types.Interface); ok {
				for i := 0; i < iface.NumMethods(); i++ {
					method := iface.Method(i)
					methodNames = append(methodNames, method.Name())
				}
			}
		}
	}

	return methodNames, nil
}

// ParseAllStructs scans a directory and returns all struct types (name -> *ast.StructType).
func ParseAllStructs(dir string) (map[string]*ast.StructType, error) {
	structs := make(map[string]*ast.StructType)

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || !strings.HasSuffix(path, ".go") {
			return nil
		}

		fset := token.NewFileSet()
		node, err := parser.ParseFile(fset, path, nil, 0)
		if err != nil {
			return err
		}

		for _, decl := range node.Decls {
			genDecl, ok := decl.(*ast.GenDecl)
			if !ok || genDecl.Tok != token.TYPE {
				continue
			}

			for _, spec := range genDecl.Specs {
				typeSpec := spec.(*ast.TypeSpec)
				structType, ok := typeSpec.Type.(*ast.StructType)
				if ok {
					structs[typeSpec.Name.Name] = structType
				}
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}
	return structs, nil
}
