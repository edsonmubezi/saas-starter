package generator

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type UsecaseTestData struct {
	ModulePath         string
	Entity             string
	EntityLower        string
	EntityPlural       string
	InterfaceName      string
	TableName          string
	NameField          string
	NameFieldGo        string
	AddInputExample    string
	UpdateInputExample string
}

type StructField struct {
	Name        string
	Type        string
	SampleValue string
}

type HandlerTestData struct {
	ModulePath     string
	Entity         string
	EntityLower    string
	EntityPlural   string
	PackageName    string
	AddInputFields string
	UpdateFields   string
	NameField      string
	NameFieldSnake string
}

type HandlerIntegrationTestData struct {
	ModulePath        string
	Entity            string
	EntityLower       string
	EntityPlural      string
	NameField         string
	NameFieldSnake    string
	CreateJSONPayload string
	UpdateJSONPayload string
}

// GenerateRepoTests generates repository integration test based on blueprint + interface methods
func GenerateRepoTests(sourcePath string, blueprintPath string) error {
	// Step 1: Parse repository method names (e.g., CreateDepartment, GetByID, etc.)
	methods, err := ParseInterfaceMethods(sourcePath)
	if err != nil {
		return fmt.Errorf("failed to parse interface: %w", err)
	}

	// Step 2: Load the test blueprint (e.g., department.json)
	blueprint, err := LoadBlueprint(blueprintPath)
	if err != nil {
		return fmt.Errorf("failed to load blueprint: %w", err)
	}

	// Step 3: Prepare output directory
	entityLower := strings.ToLower(blueprint.Entity)
	outputDir := filepath.Join("tests", entityLower)
	if err := os.MkdirAll(outputDir, os.ModePerm); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Step 4: Prepare output file
	outputFile := filepath.Join(outputDir, fmt.Sprintf("%srepository_test.go", entityLower))
	f, err := os.Create(outputFile)
	if err != nil {
		return fmt.Errorf("failed to create test file: %w", err)
	}
	defer f.Close()

	// Step 5: Render the test template
	err = RepoTestTemplate.Execute(f, map[string]any{
		"Entity":      blueprint.Entity,
		"EntityLower": entityLower,
		"CreateInput": blueprint.CreateInput,
		"Methods":     methods,
	})
	if err != nil {
		return fmt.Errorf("failed to render template: %w", err)
	}

	return nil
}

func GenerateUsecaseTests(sourcePath string) error {
	entityName := extractEntityNameFromPath(sourcePath)
	entityLower := strings.ToLower(entityName)
	entityPlural := entityName + "s"

	modulePath, err := getGoModulePath()
	if err != nil {
		return fmt.Errorf("failed to get module path: %w", err)
	}

	addStruct, updateStruct, err := ParseAddAndUpdateStructs(sourcePath)
	if err != nil {
		return fmt.Errorf("failed to parse input structs: %w", err)
	}

	addFields := StructFields(addStruct)
	updateFields := StructFields(updateStruct)

	data := UsecaseTestData{
		ModulePath:         modulePath,
		Entity:             entityName,
		EntityLower:        entityLower,
		EntityPlural:       entityPlural,
		InterfaceName:      entityName + "Repository",
		TableName:          toSnakeCase(entityPlural),
		NameField:          inferNameField(addFields),
		NameFieldGo:        inferNameField(addFields),
		AddInputExample:    formatFieldExamples(addFields),
		UpdateInputExample: formatFieldExamples(updateFields),
	}

	outputPath := filepath.Join("tests", fmt.Sprintf("%s_usecase_test.go", entityLower))
	return WriteTemplateToFile(UsecaseTestTemplate, data, outputPath)
}

func GenerateHandlerUnitTests(sourcePath string) error {
	entity := extractEntityNameForHandler(sourcePath) // e.g., "Department"
	entityLower := strings.ToLower(entity)            // e.g., "department"
	entityPlural := entity + "s"

	// Assume Add/Update structs are in: internal/<entityLower>/entity.go
	entityPackagePath := filepath.Join("internal", entityLower, "entity.go")

	addStruct, updateStruct, err := ParseAddAndUpdateStructs(entityPackagePath)
	if err != nil {
		return fmt.Errorf("failed to parse Add/Update structs: %w", err)
	}

	modulePath, err := getGoModulePath()
	if err != nil {
		return fmt.Errorf("failed to read go.mod module path: %w", err)
	}

	addFields := StructFields(addStruct)
	updateFields := StructFields(updateStruct)
	nameField := inferNameField(addFields)

	data := HandlerTestData{
		ModulePath:     modulePath,
		Entity:         entity,
		EntityLower:    entityLower,
		EntityPlural:   entityPlural,
		AddInputFields: formatFieldExamples(addFields),
		UpdateFields:   formatFieldExamples(updateFields),
		NameField:      nameField,
		NameFieldSnake: toSnakeCase(nameField),
	}

	outputDir := "tests"
	outputFile := filepath.Join(outputDir, fmt.Sprintf("%s_handler_test.go", entityLower))
	if err := os.MkdirAll(outputDir, os.ModePerm); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	return WriteTemplateToFile(HandlerUnitTestTemplate, data, outputFile)
}

func GenerateHandlerIntegrationTests(sourcePath string) error {
	entity := extractEntityNameForHandler(sourcePath)
	entityLower := strings.ToLower(entity)
	entityPlural := entity + "s"

	modulePath, err := getGoModulePath()
	if err != nil {
		return fmt.Errorf("failed to get module path: %w", err)
	}

	// Determine entity source path (e.g., internal/department)
	entitySourcePath := strings.Replace(sourcePath, "api/handler", "internal", 1)
	entitySourcePath = strings.TrimSuffix(entitySourcePath, "_handler.go")

	addStruct, err := parseAddOnly(entitySourcePath)
	if err != nil {
		return fmt.Errorf("failed to parse Add struct: %w", err)
	}

	fields := StructFields(addStruct)
	samplePayload := buildJSONPayload(fields)

	testData := HandlerIntegrationTestData{
		ModulePath:        modulePath,
		Entity:            entity,
		EntityLower:       entityLower,
		EntityPlural:      entityPlural,
		CreateJSONPayload: samplePayload,
	}

	outputDir := filepath.Join("tests")
	outputFile := filepath.Join(outputDir, fmt.Sprintf("%s_handler_integration_test.go", entityLower))
	if err := os.MkdirAll(outputDir, os.ModePerm); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	f, err := os.Create(outputFile)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer f.Close()

	if err := HandlerIntegrationTestTemplate.Execute(f, testData); err != nil {
		return fmt.Errorf("failed to render template: %w", err)
	}

	return nil
}
