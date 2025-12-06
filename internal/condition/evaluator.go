package condition

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/piwi3910/openfroyo/internal/template"
)

// Evaluator evaluates conditional expressions
type Evaluator struct {
	ctx *template.Context
}

// NewEvaluator creates a new condition evaluator
func NewEvaluator(ctx *template.Context) *Evaluator {
	return &Evaluator{ctx: ctx}
}

// Evaluate evaluates a conditional expression and returns true/false
func (e *Evaluator) Evaluate(expr string) (bool, error) {
	if expr == "" {
		return true, nil
	}

	expr = strings.TrimSpace(expr)

	// Handle logical operators (and, or)
	if result, ok, err := e.evaluateLogical(expr); ok {
		return result, err
	}

	// Handle comparison operators
	if result, ok, err := e.evaluateComparison(expr); ok {
		return result, err
	}

	// Handle 'not' prefix
	if strings.HasPrefix(expr, "not ") {
		result, err := e.Evaluate(strings.TrimPrefix(expr, "not "))
		if err != nil {
			return false, err
		}
		return !result, nil
	}

	// Handle 'in' operator
	if result, ok, err := e.evaluateIn(expr); ok {
		return result, err
	}

	// Handle defined/undefined checks
	if strings.HasPrefix(expr, "defined(") && strings.HasSuffix(expr, ")") {
		varName := strings.TrimSuffix(strings.TrimPrefix(expr, "defined("), ")")
		varName = strings.TrimSpace(varName)
		_, err := e.getValue(varName)
		return err == nil, nil
	}

	// Handle boolean literals
	if expr == "true" || expr == "True" || expr == "TRUE" {
		return true, nil
	}
	if expr == "false" || expr == "False" || expr == "FALSE" {
		return false, nil
	}

	// Try to evaluate as a single value (truthy/falsy)
	value, err := e.getValue(expr)
	if err != nil {
		return false, nil // Undefined variables are falsy
	}

	return isTruthy(value), nil
}

// evaluateLogical handles 'and' and 'or' operators
func (e *Evaluator) evaluateLogical(expr string) (bool, bool, error) {
	// Split on ' and ' (with spaces to avoid matching within words)
	if idx := strings.Index(expr, " and "); idx != -1 {
		left := strings.TrimSpace(expr[:idx])
		right := strings.TrimSpace(expr[idx+5:])

		leftResult, err := e.Evaluate(left)
		if err != nil {
			return false, true, err
		}
		if !leftResult {
			return false, true, nil // Short circuit
		}

		rightResult, err := e.Evaluate(right)
		if err != nil {
			return false, true, err
		}
		return rightResult, true, nil
	}

	// Split on ' or '
	if idx := strings.Index(expr, " or "); idx != -1 {
		left := strings.TrimSpace(expr[:idx])
		right := strings.TrimSpace(expr[idx+4:])

		leftResult, err := e.Evaluate(left)
		if err != nil {
			return false, true, err
		}
		if leftResult {
			return true, true, nil // Short circuit
		}

		rightResult, err := e.Evaluate(right)
		if err != nil {
			return false, true, err
		}
		return rightResult, true, nil
	}

	return false, false, nil
}

// evaluateComparison handles comparison operators
func (e *Evaluator) evaluateComparison(expr string) (bool, bool, error) {
	// Order matters: check longer operators first
	operators := []struct {
		op string
		fn func(left, right interface{}) (bool, error)
	}{
		{"!=", e.notEqual},
		{"==", e.equal},
		{">=", e.greaterEqual},
		{"<=", e.lessEqual},
		{">", e.greater},
		{"<", e.less},
	}

	for _, op := range operators {
		if idx := strings.Index(expr, op.op); idx != -1 {
			left := strings.TrimSpace(expr[:idx])
			right := strings.TrimSpace(expr[idx+len(op.op):])

			leftVal, err := e.getValue(left)
			if err != nil {
				return false, true, err
			}

			rightVal, err := e.getValue(right)
			if err != nil {
				return false, true, err
			}

			result, err := op.fn(leftVal, rightVal)
			return result, true, err
		}
	}

	return false, false, nil
}

// evaluateIn handles 'in' operator
func (e *Evaluator) evaluateIn(expr string) (bool, bool, error) {
	// Match "value in list" or "value not in list"
	notIn := false
	inIdx := strings.Index(expr, " not in ")
	if inIdx != -1 {
		notIn = true
	} else {
		inIdx = strings.Index(expr, " in ")
		if inIdx == -1 {
			return false, false, nil
		}
	}

	var left, right string
	if notIn {
		left = strings.TrimSpace(expr[:inIdx])
		right = strings.TrimSpace(expr[inIdx+8:])
	} else {
		left = strings.TrimSpace(expr[:inIdx])
		right = strings.TrimSpace(expr[inIdx+4:])
	}

	leftVal, err := e.getValue(left)
	if err != nil {
		return false, true, err
	}

	rightVal, err := e.getValue(right)
	if err != nil {
		return false, true, err
	}

	// Check if leftVal is in rightVal (which should be a list)
	switch rv := rightVal.(type) {
	case []interface{}:
		for _, item := range rv {
			if e.equalValues(leftVal, item) {
				return !notIn, true, nil
			}
		}
		return notIn, true, nil
	case []string:
		leftStr := fmt.Sprintf("%v", leftVal)
		for _, item := range rv {
			if leftStr == item {
				return !notIn, true, nil
			}
		}
		return notIn, true, nil
	case string:
		// Check if substring
		leftStr := fmt.Sprintf("%v", leftVal)
		contains := strings.Contains(rv, leftStr)
		if notIn {
			return !contains, true, nil
		}
		return contains, true, nil
	default:
		return false, true, fmt.Errorf("'in' operator requires a list or string on the right side")
	}
}

// getValue retrieves a value from the context or parses a literal
func (e *Evaluator) getValue(expr string) (interface{}, error) {
	expr = strings.TrimSpace(expr)

	// String literal (single or double quoted)
	if (strings.HasPrefix(expr, "'") && strings.HasSuffix(expr, "'")) ||
		(strings.HasPrefix(expr, "\"") && strings.HasSuffix(expr, "\"")) {
		return expr[1 : len(expr)-1], nil
	}

	// Numeric literal
	if num, err := strconv.ParseFloat(expr, 64); err == nil {
		// Return as int if it's a whole number
		if num == float64(int64(num)) {
			return int64(num), nil
		}
		return num, nil
	}

	// Boolean literal
	if expr == "true" || expr == "True" || expr == "TRUE" {
		return true, nil
	}
	if expr == "false" || expr == "False" || expr == "FALSE" {
		return false, nil
	}

	// nil/null literal
	if expr == "nil" || expr == "null" || expr == "none" || expr == "None" {
		return nil, nil
	}

	// List literal [a, b, c]
	if strings.HasPrefix(expr, "[") && strings.HasSuffix(expr, "]") {
		return e.parseList(expr[1 : len(expr)-1])
	}

	// Variable reference (var.name, vars.name, host.property, or registered variable)
	if strings.HasPrefix(expr, "var.") || strings.HasPrefix(expr, "vars.") {
		processed, err := template.Process("{{ "+expr+" }}", e.ctx)
		if err != nil {
			return nil, err
		}
		// If it wasn't processed (returned as-is), it's undefined
		if processed == "{{ "+expr+" }}" {
			return nil, fmt.Errorf("undefined variable: %s", expr)
		}
		// Try to parse as number
		if num, err := strconv.ParseFloat(processed, 64); err == nil {
			if num == float64(int64(num)) {
				return int64(num), nil
			}
			return num, nil
		}
		// Try to parse as boolean
		if processed == "true" {
			return true, nil
		}
		if processed == "false" {
			return false, nil
		}
		return processed, nil
	}

	// Host property
	if strings.HasPrefix(expr, "host.") {
		processed, err := template.Process("{{ "+expr+" }}", e.ctx)
		if err != nil {
			return nil, err
		}
		return processed, nil
	}

	// Registered variable (just a name, look up in Vars)
	if val, ok := e.ctx.Vars[expr]; ok {
		return val, nil
	}

	// Nested registered variable access (e.g., result.status, result.facts.key)
	if strings.Contains(expr, ".") {
		parts := strings.SplitN(expr, ".", 2)
		if val, ok := e.ctx.Vars[parts[0]]; ok {
			return e.getNestedValue(val, parts[1])
		}
	}

	return nil, fmt.Errorf("undefined: %s", expr)
}

// getNestedValue gets a nested value from a map or struct-like value
func (e *Evaluator) getNestedValue(value interface{}, path string) (interface{}, error) {
	parts := strings.Split(path, ".")
	current := value

	for _, part := range parts {
		switch v := current.(type) {
		case map[string]interface{}:
			val, ok := v[part]
			if !ok {
				return nil, fmt.Errorf("key not found: %s", part)
			}
			current = val
		case map[string]string:
			val, ok := v[part]
			if !ok {
				return nil, fmt.Errorf("key not found: %s", part)
			}
			current = val
		default:
			return nil, fmt.Errorf("cannot access %s on type %T", part, current)
		}
	}

	return current, nil
}

// parseList parses a list literal
func (e *Evaluator) parseList(content string) ([]interface{}, error) {
	var result []interface{}
	items := splitListItems(content)

	for _, item := range items {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}
		val, err := e.getValue(item)
		if err != nil {
			// Treat as string if can't parse
			result = append(result, item)
		} else {
			result = append(result, val)
		}
	}

	return result, nil
}

// splitListItems splits list items handling nested structures
func splitListItems(content string) []string {
	var items []string
	var current strings.Builder
	depth := 0
	inQuote := false
	quoteChar := rune(0)

	for _, r := range content {
		switch {
		case (r == '\'' || r == '"') && !inQuote:
			inQuote = true
			quoteChar = r
			current.WriteRune(r)
		case r == quoteChar && inQuote:
			inQuote = false
			quoteChar = 0
			current.WriteRune(r)
		case r == '[' || r == '{':
			depth++
			current.WriteRune(r)
		case r == ']' || r == '}':
			depth--
			current.WriteRune(r)
		case r == ',' && depth == 0 && !inQuote:
			items = append(items, current.String())
			current.Reset()
		default:
			current.WriteRune(r)
		}
	}

	if current.Len() > 0 {
		items = append(items, current.String())
	}

	return items
}

// Comparison functions
func (e *Evaluator) equal(left, right interface{}) (bool, error) {
	return e.equalValues(left, right), nil
}

func (e *Evaluator) notEqual(left, right interface{}) (bool, error) {
	return !e.equalValues(left, right), nil
}

func (e *Evaluator) equalValues(left, right interface{}) bool {
	// Handle nil
	if left == nil && right == nil {
		return true
	}
	if left == nil || right == nil {
		return false
	}

	// Convert to comparable types
	leftStr := fmt.Sprintf("%v", left)
	rightStr := fmt.Sprintf("%v", right)

	// Try numeric comparison first
	leftNum, leftErr := strconv.ParseFloat(leftStr, 64)
	rightNum, rightErr := strconv.ParseFloat(rightStr, 64)
	if leftErr == nil && rightErr == nil {
		return leftNum == rightNum
	}

	// String comparison
	return leftStr == rightStr
}

func (e *Evaluator) greater(left, right interface{}) (bool, error) {
	leftNum, rightNum, err := e.toNumbers(left, right)
	if err != nil {
		// Fall back to string comparison
		return fmt.Sprintf("%v", left) > fmt.Sprintf("%v", right), nil
	}
	return leftNum > rightNum, nil
}

func (e *Evaluator) less(left, right interface{}) (bool, error) {
	leftNum, rightNum, err := e.toNumbers(left, right)
	if err != nil {
		return fmt.Sprintf("%v", left) < fmt.Sprintf("%v", right), nil
	}
	return leftNum < rightNum, nil
}

func (e *Evaluator) greaterEqual(left, right interface{}) (bool, error) {
	leftNum, rightNum, err := e.toNumbers(left, right)
	if err != nil {
		return fmt.Sprintf("%v", left) >= fmt.Sprintf("%v", right), nil
	}
	return leftNum >= rightNum, nil
}

func (e *Evaluator) lessEqual(left, right interface{}) (bool, error) {
	leftNum, rightNum, err := e.toNumbers(left, right)
	if err != nil {
		return fmt.Sprintf("%v", left) <= fmt.Sprintf("%v", right), nil
	}
	return leftNum <= rightNum, nil
}

func (e *Evaluator) toNumbers(left, right interface{}) (float64, float64, error) {
	leftStr := fmt.Sprintf("%v", left)
	rightStr := fmt.Sprintf("%v", right)

	leftNum, err := strconv.ParseFloat(leftStr, 64)
	if err != nil {
		return 0, 0, err
	}

	rightNum, err := strconv.ParseFloat(rightStr, 64)
	if err != nil {
		return 0, 0, err
	}

	return leftNum, rightNum, nil
}

// isTruthy determines if a value is truthy
func isTruthy(value interface{}) bool {
	if value == nil {
		return false
	}

	switch v := value.(type) {
	case bool:
		return v
	case int, int8, int16, int32, int64:
		return v != 0
	case uint, uint8, uint16, uint32, uint64:
		return v != 0
	case float32:
		return v != 0
	case float64:
		return v != 0
	case string:
		return v != "" && v != "false" && v != "False" && v != "FALSE"
	case []interface{}:
		return len(v) > 0
	case map[string]interface{}:
		return len(v) > 0
	default:
		return true
	}
}

// EvaluateWithContext evaluates a conditional expression with the given context
func EvaluateWithContext(expr string, vars map[string]interface{}, host template.HostContext) (bool, error) {
	ctx := &template.Context{
		Vars: vars,
		Host: host,
	}
	evaluator := NewEvaluator(ctx)
	return evaluator.Evaluate(expr)
}

// Common patterns for expression matching
var (
	// Matches registered variable access like: result.status, result.facts.key
	registeredVarPattern = regexp.MustCompile(`^([a-zA-Z_][a-zA-Z0-9_]*)\.(.+)$`)
)
