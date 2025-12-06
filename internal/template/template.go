package template

import (
	"fmt"
	"regexp"
	"strings"
)

// Context provides data for template processing
type Context struct {
	Vars map[string]interface{}
	Host HostContext
}

// HostContext provides host-specific data for templates
type HostContext struct {
	Name       string
	SSHHost    string
	SSHPort    int
	SSHUser    string
	Datacenter string
	AgentID    string
	Mode       string
}

// templatePattern matches {{ expression }} syntax
var templatePattern = regexp.MustCompile(`\{\{\s*(.+?)\s*\}\}`)

// Process processes template expressions in a string
func Process(input string, ctx *Context) (string, error) {
	result := templatePattern.ReplaceAllStringFunc(input, func(match string) string {
		// Extract expression from {{ }}
		expr := templatePattern.FindStringSubmatch(match)[1]
		value, err := evaluateExpression(expr, ctx)
		if err != nil {
			return match // Return original on error
		}
		return fmt.Sprintf("%v", value)
	})
	return result, nil
}

// ProcessVars processes template expressions in all variable values
func ProcessVars(vars map[string]interface{}, ctx *Context) (map[string]interface{}, error) {
	result := make(map[string]interface{})
	for k, v := range vars {
		processed, err := processValue(v, ctx)
		if err != nil {
			return nil, fmt.Errorf("error processing variable %s: %w", k, err)
		}
		result[k] = processed
	}
	return result, nil
}

// processValue recursively processes template expressions in a value
func processValue(value interface{}, ctx *Context) (interface{}, error) {
	switch v := value.(type) {
	case string:
		return Process(v, ctx)
	case map[string]interface{}:
		result := make(map[string]interface{})
		for k, val := range v {
			processed, err := processValue(val, ctx)
			if err != nil {
				return nil, err
			}
			result[k] = processed
		}
		return result, nil
	case []interface{}:
		result := make([]interface{}, len(v))
		for i, val := range v {
			processed, err := processValue(val, ctx)
			if err != nil {
				return nil, err
			}
			result[i] = processed
		}
		return result, nil
	default:
		return value, nil
	}
}

// evaluateExpression evaluates a template expression
func evaluateExpression(expr string, ctx *Context) (interface{}, error) {
	expr = strings.TrimSpace(expr)

	// Check for filter (pipe) syntax: expression | filter(args)
	if idx := strings.Index(expr, "|"); idx != -1 {
		baseExpr := strings.TrimSpace(expr[:idx])
		filterExpr := strings.TrimSpace(expr[idx+1:])
		baseValue, err := evaluateExpression(baseExpr, ctx)
		if err != nil {
			return nil, err
		}
		return applyFilter(baseValue, filterExpr, ctx)
	}

	// Variable reference: var.name or vars.name
	if strings.HasPrefix(expr, "var.") || strings.HasPrefix(expr, "vars.") {
		varName := strings.TrimPrefix(strings.TrimPrefix(expr, "vars."), "var.")
		return getNestedValue(ctx.Vars, varName)
	}

	// Host reference: host.property
	if strings.HasPrefix(expr, "host.") {
		propName := strings.TrimPrefix(expr, "host.")
		return getHostProperty(ctx.Host, propName)
	}

	// Direct variable lookup (backward compatibility)
	if val, ok := ctx.Vars[expr]; ok {
		return val, nil
	}

	return nil, fmt.Errorf("unknown expression: %s", expr)
}

// getNestedValue retrieves a nested value from a map using dot notation
func getNestedValue(data map[string]interface{}, path string) (interface{}, error) {
	parts := strings.Split(path, ".")
	current := interface{}(data)

	for _, part := range parts {
		switch v := current.(type) {
		case map[string]interface{}:
			val, ok := v[part]
			if !ok {
				return nil, fmt.Errorf("key not found: %s", part)
			}
			current = val
		default:
			return nil, fmt.Errorf("cannot access %s on non-map type", part)
		}
	}

	return current, nil
}

// getHostProperty retrieves a host property by name
func getHostProperty(host HostContext, prop string) (interface{}, error) {
	switch prop {
	case "name":
		return host.Name, nil
	case "ssh_host":
		return host.SSHHost, nil
	case "ssh_port":
		return host.SSHPort, nil
	case "ssh_user":
		return host.SSHUser, nil
	case "datacenter":
		return host.Datacenter, nil
	case "agent_id":
		return host.AgentID, nil
	case "mode":
		return host.Mode, nil
	default:
		return nil, fmt.Errorf("unknown host property: %s", prop)
	}
}

// applyFilter applies a filter function to a value
func applyFilter(value interface{}, filterExpr string, ctx *Context) (interface{}, error) {
	// Parse filter name and arguments
	filterName := filterExpr
	var args []string

	if idx := strings.Index(filterExpr, "("); idx != -1 {
		filterName = strings.TrimSpace(filterExpr[:idx])
		argsStr := strings.TrimSuffix(strings.TrimSpace(filterExpr[idx+1:]), ")")
		if argsStr != "" {
			args = parseArgs(argsStr)
		}
	}

	switch filterName {
	case "default":
		if value == nil || value == "" {
			if len(args) > 0 {
				return args[0], nil
			}
			return "", nil
		}
		return value, nil

	case "upper":
		if s, ok := value.(string); ok {
			return strings.ToUpper(s), nil
		}
		return value, nil

	case "lower":
		if s, ok := value.(string); ok {
			return strings.ToLower(s), nil
		}
		return value, nil

	case "trim":
		if s, ok := value.(string); ok {
			return strings.TrimSpace(s), nil
		}
		return value, nil

	case "replace":
		if s, ok := value.(string); ok && len(args) >= 2 {
			return strings.ReplaceAll(s, args[0], args[1]), nil
		}
		return value, nil

	case "split":
		if s, ok := value.(string); ok && len(args) >= 1 {
			return strings.Split(s, args[0]), nil
		}
		return value, nil

	case "join":
		if arr, ok := value.([]interface{}); ok && len(args) >= 1 {
			strs := make([]string, len(arr))
			for i, v := range arr {
				strs[i] = fmt.Sprintf("%v", v)
			}
			return strings.Join(strs, args[0]), nil
		}
		if arr, ok := value.([]string); ok && len(args) >= 1 {
			return strings.Join(arr, args[0]), nil
		}
		return value, nil

	case "first":
		if arr, ok := value.([]interface{}); ok && len(arr) > 0 {
			return arr[0], nil
		}
		return nil, nil

	case "last":
		if arr, ok := value.([]interface{}); ok && len(arr) > 0 {
			return arr[len(arr)-1], nil
		}
		return nil, nil

	case "length":
		switch v := value.(type) {
		case string:
			return len(v), nil
		case []interface{}:
			return len(v), nil
		case map[string]interface{}:
			return len(v), nil
		}
		return 0, nil

	default:
		return nil, fmt.Errorf("unknown filter: %s", filterName)
	}
}

// parseArgs parses comma-separated arguments, handling quoted strings
func parseArgs(argsStr string) []string {
	var args []string
	var current strings.Builder
	inQuote := false
	quoteChar := rune(0)

	for _, r := range argsStr {
		switch {
		case (r == '\'' || r == '"') && !inQuote:
			inQuote = true
			quoteChar = r
		case r == quoteChar && inQuote:
			inQuote = false
			quoteChar = 0
		case r == ',' && !inQuote:
			args = append(args, strings.TrimSpace(current.String()))
			current.Reset()
		default:
			current.WriteRune(r)
		}
	}

	if current.Len() > 0 {
		args = append(args, strings.TrimSpace(current.String()))
	}

	return args
}
