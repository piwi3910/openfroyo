package validator

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/piwi3910/openfroyo/internal/parser"
)

// ValidationError represents a validation error with location info
type ValidationError struct {
	Type    string // "syntax", "schema", "module", "inventory", "variable"
	Message string
	Line    int // Optional line number
}

func (e ValidationError) Error() string {
	if e.Line > 0 {
		return fmt.Sprintf("[%s] Line %d: %s", e.Type, e.Line, e.Message)
	}
	return fmt.Sprintf("[%s] %s", e.Type, e.Message)
}

// ValidationResult contains all validation results
type ValidationResult struct {
	Errors   []ValidationError
	Warnings []ValidationError
	Valid    bool
}

// Validator validates stack files
type Validator struct {
	moduleDir string
}

// NewValidator creates a new validator
func NewValidator(moduleDir string) *Validator {
	return &Validator{
		moduleDir: moduleDir,
	}
}

// Validate performs full validation on a stack file
func (v *Validator) Validate(stackPath string) (*ValidationResult, error) {
	result := &ValidationResult{
		Errors:   []ValidationError{},
		Warnings: []ValidationError{},
		Valid:    true,
	}

	// Step 1: Check if file exists
	if _, err := os.Stat(stackPath); os.IsNotExist(err) {
		result.Errors = append(result.Errors, ValidationError{
			Type:    "syntax",
			Message: fmt.Sprintf("Stack file not found: %s", stackPath),
		})
		result.Valid = false
		return result, nil
	}

	// Step 2: Load and parse stack (validates YAML syntax)
	stack, err := parser.LoadStack(stackPath)
	if err != nil {
		result.Errors = append(result.Errors, ValidationError{
			Type:    "syntax",
			Message: fmt.Sprintf("Failed to parse stack: %v", err),
		})
		result.Valid = false
		return result, nil
	}

	// Step 3: Schema validation
	v.validateSchema(stack, result)

	// Step 4: Load inventory and validate
	var inventory *parser.Inventory
	if len(stack.Inventory) > 0 {
		inventory, err = parser.LoadInventory(stack.Inventory)
		if err != nil {
			result.Errors = append(result.Errors, ValidationError{
				Type:    "inventory",
				Message: fmt.Sprintf("Failed to load inventory: %v", err),
			})
			result.Valid = false
		}
	} else {
		result.Errors = append(result.Errors, ValidationError{
			Type:    "schema",
			Message: "No inventory specified in stack",
		})
		result.Valid = false
	}

	// Step 5: Validate modules exist
	v.validateModules(stack, result)

	// Step 6: Validate host references
	if inventory != nil {
		v.validateHosts(stack, inventory, result)
	}

	// Step 7: Validate handlers
	v.validateHandlers(stack, result)

	// Step 8: Check for common issues (warnings)
	v.checkCommonIssues(stack, result)

	// Update valid flag
	result.Valid = len(result.Errors) == 0

	return result, nil
}

// validateSchema checks for required fields
func (v *Validator) validateSchema(stack *parser.Stack, result *ValidationResult) {
	if stack.Name == "" {
		result.Errors = append(result.Errors, ValidationError{
			Type:    "schema",
			Message: "Stack name is required",
		})
	}

	if len(stack.Run) == 0 {
		result.Errors = append(result.Errors, ValidationError{
			Type:    "schema",
			Message: "Stack has no run entries",
		})
	}

	for i, entry := range stack.Run {
		if entry.Name == "" {
			result.Errors = append(result.Errors, ValidationError{
				Type:    "schema",
				Message: fmt.Sprintf("Run entry %d has no name", i+1),
			})
		}

		// Must have either module or tasks
		if entry.Module == "" && len(entry.Tasks) == 0 {
			result.Errors = append(result.Errors, ValidationError{
				Type:    "schema",
				Message: fmt.Sprintf("Run entry '%s' must have either 'module' or 'tasks'", entry.Name),
			})
		}

		// Validate inline tasks
		for j, task := range entry.Tasks {
			if task.Name == "" {
				result.Errors = append(result.Errors, ValidationError{
					Type:    "schema",
					Message: fmt.Sprintf("Task %d in entry '%s' has no name", j+1, entry.Name),
				})
			}
			if task.Module == "" {
				result.Errors = append(result.Errors, ValidationError{
					Type:    "schema",
					Message: fmt.Sprintf("Task '%s' in entry '%s' has no module", task.Name, entry.Name),
				})
			}
		}
	}
}

// validateModules checks if referenced modules exist
func (v *Validator) validateModules(stack *parser.Stack, result *ValidationResult) {
	checkedModules := make(map[string]bool)

	for _, entry := range stack.Run {
		// Check entry module
		if entry.Module != "" {
			if !checkedModules[entry.Module] {
				if !v.moduleExists(entry.Module) {
					result.Errors = append(result.Errors, ValidationError{
						Type:    "module",
						Message: fmt.Sprintf("Module not found: %s", entry.Module),
					})
				}
				checkedModules[entry.Module] = true
			}
		}

		// Check task modules
		for _, task := range entry.Tasks {
			if task.Module != "" && !checkedModules[task.Module] {
				if !v.moduleExists(task.Module) {
					result.Errors = append(result.Errors, ValidationError{
						Type:    "module",
						Message: fmt.Sprintf("Module not found: %s (in task '%s')", task.Module, task.Name),
					})
				}
				checkedModules[task.Module] = true
			}
		}
	}

	// Check handler modules
	for _, handler := range stack.Handlers {
		if handler.Module != "" && !checkedModules[handler.Module] {
			if !v.moduleExists(handler.Module) {
				result.Errors = append(result.Errors, ValidationError{
					Type:    "module",
					Message: fmt.Sprintf("Module not found: %s (in handler '%s')", handler.Module, handler.Name),
				})
			}
			checkedModules[handler.Module] = true
		}
	}
}

// moduleExists checks if a module WASM file exists
func (v *Validator) moduleExists(moduleName string) bool {
	moduleNameParts := strings.Split(moduleName, "/")
	wasmName := moduleNameParts[len(moduleNameParts)-1]
	modulePath := filepath.Join(v.moduleDir, moduleName, "wasm", wasmName+".wasm")
	_, err := os.Stat(modulePath)
	return err == nil
}

// validateHosts checks if all referenced hosts/groups exist in inventory
func (v *Validator) validateHosts(stack *parser.Stack, inventory *parser.Inventory, result *ValidationResult) {
	for _, entry := range stack.Run {
		for _, hostRef := range entry.Hosts {
			if err := v.validateHostRef(hostRef, inventory); err != nil {
				result.Errors = append(result.Errors, ValidationError{
					Type:    "inventory",
					Message: fmt.Sprintf("In entry '%s': %s", entry.Name, err.Error()),
				})
			}
		}
	}

	// Check handler hosts
	for _, handler := range stack.Handlers {
		for _, hostRef := range handler.Hosts {
			if err := v.validateHostRef(hostRef, inventory); err != nil {
				result.Errors = append(result.Errors, ValidationError{
					Type:    "inventory",
					Message: fmt.Sprintf("In handler '%s': %s", handler.Name, err.Error()),
				})
			}
		}
	}
}

// validateHostRef validates a single host reference
func (v *Validator) validateHostRef(hostRef string, inventory *parser.Inventory) error {
	if strings.HasPrefix(hostRef, "@group:") {
		groupName := strings.TrimPrefix(hostRef, "@group:")
		if _, ok := inventory.Groups[groupName]; !ok {
			return fmt.Errorf("group not found: %s", groupName)
		}
	} else {
		if _, ok := inventory.Hosts[hostRef]; !ok {
			return fmt.Errorf("host not found: %s", hostRef)
		}
	}
	return nil
}

// validateHandlers checks handler definitions and notify references
func (v *Validator) validateHandlers(stack *parser.Stack, result *ValidationResult) {
	// Build map of defined handlers
	definedHandlers := make(map[string]bool)
	for _, handler := range stack.Handlers {
		if handler.Name == "" {
			result.Errors = append(result.Errors, ValidationError{
				Type:    "schema",
				Message: "Handler has no name",
			})
			continue
		}
		if definedHandlers[handler.Name] {
			result.Warnings = append(result.Warnings, ValidationError{
				Type:    "schema",
				Message: fmt.Sprintf("Duplicate handler name: %s", handler.Name),
			})
		}
		definedHandlers[handler.Name] = true
	}

	// Check all notify references
	for _, entry := range stack.Run {
		for _, notifyName := range entry.Notify {
			if !definedHandlers[notifyName] {
				result.Errors = append(result.Errors, ValidationError{
					Type:    "schema",
					Message: fmt.Sprintf("Entry '%s' notifies undefined handler: %s", entry.Name, notifyName),
				})
			}
		}

		// Check task notify references
		for _, task := range entry.Tasks {
			for _, notifyName := range task.Notify {
				if !definedHandlers[notifyName] {
					result.Errors = append(result.Errors, ValidationError{
						Type:    "schema",
						Message: fmt.Sprintf("Task '%s' notifies undefined handler: %s", task.Name, notifyName),
					})
				}
			}
		}
	}
}

// checkCommonIssues looks for potential problems that might not be errors
func (v *Validator) checkCommonIssues(stack *parser.Stack, result *ValidationResult) {
	// Check for entries without hosts
	for _, entry := range stack.Run {
		if len(entry.Hosts) == 0 && len(entry.Tasks) == 0 {
			result.Warnings = append(result.Warnings, ValidationError{
				Type:    "schema",
				Message: fmt.Sprintf("Entry '%s' has no hosts specified", entry.Name),
			})
		}
	}

	// Check for handlers without hosts
	for _, handler := range stack.Handlers {
		if len(handler.Hosts) == 0 {
			result.Warnings = append(result.Warnings, ValidationError{
				Type:    "schema",
				Message: fmt.Sprintf("Handler '%s' has no hosts specified (will be skipped)", handler.Name),
			})
		}
	}

	// Check for unused handlers
	usedHandlers := make(map[string]bool)
	for _, entry := range stack.Run {
		for _, h := range entry.Notify {
			usedHandlers[h] = true
		}
		for _, task := range entry.Tasks {
			for _, h := range task.Notify {
				usedHandlers[h] = true
			}
		}
	}
	for _, handler := range stack.Handlers {
		if !usedHandlers[handler.Name] {
			result.Warnings = append(result.Warnings, ValidationError{
				Type:    "schema",
				Message: fmt.Sprintf("Handler '%s' is defined but never notified", handler.Name),
			})
		}
	}

	// Check for potential serial execution bottlenecks
	for _, entry := range stack.Run {
		if len(entry.Hosts) > 10 && (entry.Strategy == "" || entry.Strategy == "serial") {
			result.Warnings = append(result.Warnings, ValidationError{
				Type:    "performance",
				Message: fmt.Sprintf("Entry '%s' targets %d hosts with serial strategy - consider using 'strategy: parallel'", entry.Name, len(entry.Hosts)),
			})
		}
	}
}

// PrintResult prints the validation result in a human-readable format
func PrintResult(result *ValidationResult, stackPath string) {
	fmt.Printf("Validating: %s\n", stackPath)
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println()

	if result.Valid {
		fmt.Println("✓ Syntax: Valid YAML")
		fmt.Println("✓ Schema: All required fields present")
		fmt.Println("✓ Modules: All referenced modules found")
		fmt.Println("✓ Inventory: All hosts/groups defined")
		fmt.Println("✓ Handlers: All notify references valid")
	} else {
		// Group errors by type
		syntaxErrors := filterByType(result.Errors, "syntax")
		schemaErrors := filterByType(result.Errors, "schema")
		moduleErrors := filterByType(result.Errors, "module")
		inventoryErrors := filterByType(result.Errors, "inventory")

		if len(syntaxErrors) == 0 {
			fmt.Println("✓ Syntax: Valid YAML")
		} else {
			fmt.Println("✗ Syntax errors:")
			for _, e := range syntaxErrors {
				fmt.Printf("  - %s\n", e.Message)
			}
		}

		if len(schemaErrors) == 0 {
			fmt.Println("✓ Schema: All required fields present")
		} else {
			fmt.Println("✗ Schema errors:")
			for _, e := range schemaErrors {
				fmt.Printf("  - %s\n", e.Message)
			}
		}

		if len(moduleErrors) == 0 {
			fmt.Println("✓ Modules: All referenced modules found")
		} else {
			fmt.Println("✗ Module errors:")
			for _, e := range moduleErrors {
				fmt.Printf("  - %s\n", e.Message)
			}
		}

		if len(inventoryErrors) == 0 {
			fmt.Println("✓ Inventory: All hosts/groups defined")
		} else {
			fmt.Println("✗ Inventory errors:")
			for _, e := range inventoryErrors {
				fmt.Printf("  - %s\n", e.Message)
			}
		}
	}

	// Print warnings
	if len(result.Warnings) > 0 {
		fmt.Println()
		fmt.Println("Warnings:")
		for _, w := range result.Warnings {
			fmt.Printf("  ⚠ %s\n", w.Message)
		}
	}

	fmt.Println()
	fmt.Println(strings.Repeat("=", 60))

	if result.Valid {
		fmt.Printf("✓ Stack validation passed!\n")
	} else {
		fmt.Printf("✗ Stack validation failed with %d error(s)\n", len(result.Errors))
	}
}

func filterByType(errors []ValidationError, errType string) []ValidationError {
	var filtered []ValidationError
	for _, e := range errors {
		if e.Type == errType {
			filtered = append(filtered, e)
		}
	}
	return filtered
}
