package config

import (
	"context"
	"fmt"
	"time"

	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
)

// StarlarkEvaluator executes Starlark scripts safely.
type StarlarkEvaluator struct {
	timeout time.Duration
}

// NewStarlarkEvaluator creates a new Starlark evaluator.
func NewStarlarkEvaluator(timeout time.Duration) *StarlarkEvaluator {
	if timeout == 0 {
		timeout = 30 * time.Second // Default timeout
	}
	return &StarlarkEvaluator{
		timeout: timeout,
	}
}

// Evaluate executes a Starlark script with the given input and returns the result.
func (se *StarlarkEvaluator) Evaluate(ctx context.Context, script string, input map[string]interface{}) (*StarlarkResult, error) {
	startTime := time.Now()

	// Create timeout context
	evalCtx, cancel := context.WithTimeout(ctx, se.timeout)
	defer cancel()

	// Create channel to receive result or error
	resultCh := make(chan *StarlarkResult, 1)
	errCh := make(chan error, 1)

	go func() {
		result, err := se.evaluateSync(script, input)
		if err != nil {
			errCh <- err
		} else {
			resultCh <- result
		}
	}()

	// Wait for result or timeout
	select {
	case <-evalCtx.Done():
		return &StarlarkResult{
			ExecutionTime: time.Since(startTime),
			Error:         fmt.Sprintf("execution timeout after %v", se.timeout),
		}, fmt.Errorf("starlark execution timeout")
	case err := <-errCh:
		return &StarlarkResult{
			ExecutionTime: time.Since(startTime),
			Error:         err.Error(),
		}, err
	case result := <-resultCh:
		result.ExecutionTime = time.Since(startTime)
		return result, nil
	}
}

// evaluateSync performs the actual Starlark evaluation synchronously.
func (se *StarlarkEvaluator) evaluateSync(script string, input map[string]interface{}) (*StarlarkResult, error) {
	// Create thread
	thread := &starlark.Thread{
		Name: "openfroyo",
		Print: func(_ *starlark.Thread, msg string) {
			// Suppress print for security
		},
	}

	// Build predeclared environment with built-in functions and input
	predeclared := starlark.StringDict{
		"struct": starlarkstruct.Default,
	}

	// Add built-in helper functions
	predeclared["range"] = starlark.NewBuiltin("range", builtinRange)
	predeclared["enumerate"] = starlark.NewBuiltin("enumerate", builtinEnumerate)
	predeclared["zip"] = starlark.NewBuiltin("zip", builtinZip)

	// Convert input to Starlark values and add to predeclared
	for key, val := range input {
		starlarkVal, err := toStarlarkValue(val)
		if err != nil {
			return nil, fmt.Errorf("failed to convert input %s: %w", key, err)
		}
		predeclared[key] = starlarkVal
	}

	// Execute the script
	globals, err := starlark.ExecFile(thread, "config.star", script, predeclared)
	if err != nil {
		return nil, fmt.Errorf("starlark execution failed: %w", err)
	}

	// Convert globals to output map
	output := make(map[string]interface{})
	for name, val := range globals {
		// Skip internal variables (starting with _)
		if len(name) > 0 && name[0] == '_' {
			continue
		}
		goVal, err := fromStarlarkValue(val)
		if err != nil {
			return nil, fmt.Errorf("failed to convert output %s: %w", name, err)
		}
		output[name] = goVal
	}

	return &StarlarkResult{
		Output: output,
	}, nil
}

// toStarlarkValue converts a Go value to a Starlark value.
func toStarlarkValue(v interface{}) (starlark.Value, error) {
	if v == nil {
		return starlark.None, nil
	}

	switch val := v.(type) {
	case bool:
		return starlark.Bool(val), nil
	case int:
		return starlark.MakeInt(val), nil
	case int64:
		return starlark.MakeInt64(val), nil
	case float64:
		return starlark.Float(val), nil
	case string:
		return starlark.String(val), nil
	case []interface{}:
		list := make([]starlark.Value, len(val))
		for i, item := range val {
			starlarkItem, err := toStarlarkValue(item)
			if err != nil {
				return nil, err
			}
			list[i] = starlarkItem
		}
		return starlark.NewList(list), nil
	case map[string]interface{}:
		dict := starlark.NewDict(len(val))
		for k, v := range val {
			starlarkVal, err := toStarlarkValue(v)
			if err != nil {
				return nil, err
			}
			if err := dict.SetKey(starlark.String(k), starlarkVal); err != nil {
				return nil, err
			}
		}
		return dict, nil
	default:
		return nil, fmt.Errorf("unsupported type: %T", v)
	}
}

// fromStarlarkValue converts a Starlark value to a Go value.
func fromStarlarkValue(v starlark.Value) (interface{}, error) {
	switch val := v.(type) {
	case starlark.NoneType:
		return nil, nil
	case starlark.Bool:
		return bool(val), nil
	case starlark.Int:
		i, ok := val.Int64()
		if !ok {
			return nil, fmt.Errorf("integer too large")
		}
		return i, nil
	case starlark.Float:
		return float64(val), nil
	case starlark.String:
		return string(val), nil
	case *starlark.List:
		list := make([]interface{}, val.Len())
		for i := 0; i < val.Len(); i++ {
			item, err := fromStarlarkValue(val.Index(i))
			if err != nil {
				return nil, err
			}
			list[i] = item
		}
		return list, nil
	case *starlark.Dict:
		dict := make(map[string]interface{})
		for _, item := range val.Items() {
			key, ok := item[0].(starlark.String)
			if !ok {
				return nil, fmt.Errorf("dict key must be string")
			}
			value, err := fromStarlarkValue(item[1])
			if err != nil {
				return nil, err
			}
			dict[string(key)] = value
		}
		return dict, nil
	case *starlarkstruct.Struct:
		dict := make(map[string]interface{})
		for _, name := range val.AttrNames() {
			attr, err := val.Attr(name)
			if err != nil {
				continue
			}
			value, err := fromStarlarkValue(attr)
			if err != nil {
				return nil, err
			}
			dict[name] = value
		}
		return dict, nil
	default:
		return nil, fmt.Errorf("unsupported starlark type: %s", v.Type())
	}
}

// Built-in Starlark functions

// builtinRange implements the range() built-in function.
func builtinRange(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var start, stop, step int64 = 0, 0, 1

	switch len(args) {
	case 1:
		if err := starlark.UnpackArgs(b.Name(), args, kwargs, "stop", &stop); err != nil {
			return nil, err
		}
	case 2:
		if err := starlark.UnpackArgs(b.Name(), args, kwargs, "start", &start, "stop", &stop); err != nil {
			return nil, err
		}
	case 3:
		if err := starlark.UnpackArgs(b.Name(), args, kwargs, "start", &start, "stop", &stop, "step", &step); err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("range takes 1 to 3 arguments, got %d", len(args))
	}

	if step == 0 {
		return nil, fmt.Errorf("range step cannot be zero")
	}

	var list []starlark.Value
	if step > 0 {
		for i := start; i < stop; i += step {
			list = append(list, starlark.MakeInt64(i))
		}
	} else {
		for i := start; i > stop; i += step {
			list = append(list, starlark.MakeInt64(i))
		}
	}

	return starlark.NewList(list), nil
}

// builtinEnumerate implements the enumerate() built-in function.
func builtinEnumerate(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var iterable starlark.Iterable
	var start int64 = 0

	if err := starlark.UnpackArgs(b.Name(), args, kwargs, "iterable", &iterable, "start?", &start); err != nil {
		return nil, err
	}

	iter := iterable.Iterate()
	defer iter.Done()

	var list []starlark.Value
	var x starlark.Value
	i := start
	for iter.Next(&x) {
		tuple := starlark.Tuple{starlark.MakeInt64(i), x}
		list = append(list, tuple)
		i++
	}

	return starlark.NewList(list), nil
}

// builtinZip implements the zip() built-in function.
func builtinZip(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	if len(args) == 0 {
		return starlark.NewList(nil), nil
	}

	// Get iterators for all arguments
	iters := make([]starlark.Iterator, len(args))
	for i, arg := range args {
		iterable, ok := arg.(starlark.Iterable)
		if !ok {
			return nil, fmt.Errorf("zip argument %d is not iterable", i)
		}
		iters[i] = iterable.Iterate()
		defer iters[i].Done()
	}

	// Zip the iterables
	var list []starlark.Value
	for {
		tuple := make(starlark.Tuple, len(iters))
		for i, iter := range iters {
			if !iter.Next(&tuple[i]) {
				// One iterator is exhausted, stop
				return starlark.NewList(list), nil
			}
		}
		list = append(list, tuple)
	}
}
