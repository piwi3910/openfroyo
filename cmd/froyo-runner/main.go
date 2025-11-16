package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/imports/wasi_snapshot_preview1"
)

// TaskInput represents the input to a WASM module
type TaskInput struct {
	Vars    map[string]interface{} `json:"vars"`
	Context TaskContext            `json:"context"`
}

// TaskContext provides execution context to the module
type TaskContext struct {
	Host     string `json:"host"`
	TaskName string `json:"task_name"`
}

// TaskOutput represents the output from a WASM module
type TaskOutput struct {
	Status  string                 `json:"status"`  // "ok", "changed", "failed"
	Message string                 `json:"message"` // Human-readable message
	Facts   map[string]interface{} `json:"facts"`   // Discovered facts
}

func main() {
	var modulePath string
	var inputBase64 string

	flag.StringVar(&modulePath, "module", "", "Path to WASM module")
	flag.StringVar(&inputBase64, "input-base64", "", "Base64-encoded JSON input")
	flag.Parse()

	if modulePath == "" || inputBase64 == "" {
		fmt.Fprintln(os.Stderr, "Usage: froyo-runner --module <path.wasm> --input-base64 <base64>")
		os.Exit(1)
	}

	// Decode input
	inputJSON, err := base64.StdEncoding.DecodeString(inputBase64)
	if err != nil {
		output := TaskOutput{
			Status:  "failed",
			Message: fmt.Sprintf("Failed to decode input: %v", err),
			Facts:   make(map[string]interface{}),
		}
		printOutput(output)
		os.Exit(1)
	}

	var input TaskInput
	if err := json.Unmarshal(inputJSON, &input); err != nil {
		output := TaskOutput{
			Status:  "failed",
			Message: fmt.Sprintf("Failed to parse input JSON: %v", err),
			Facts:   make(map[string]interface{}),
		}
		printOutput(output)
		os.Exit(1)
	}

	// Read WASM module
	wasmBytes, err := os.ReadFile(modulePath)
	if err != nil {
		output := TaskOutput{
			Status:  "failed",
			Message: fmt.Sprintf("Failed to read WASM module: %v", err),
			Facts:   make(map[string]interface{}),
		}
		printOutput(output)
		os.Exit(1)
	}

	// Execute WASM module
	output := executeWASM(wasmBytes, input)
	printOutput(output)

	if output.Status == "failed" {
		os.Exit(1)
	}
}

func executeWASM(wasmBytes []byte, input TaskInput) TaskOutput {
	ctx := context.Background()

	// Create a new WebAssembly runtime
	r := wazero.NewRuntime(ctx)
	defer r.Close(ctx)

	// Instantiate WASI, which provides host functions for I/O, env vars, etc.
	wasi_snapshot_preview1.Instantiate(ctx, r)

	// Create stdin with input JSON
	inputJSON, _ := json.Marshal(input)
	stdin := strings.NewReader(string(inputJSON))

	// Capture stdout
	var stdout, stderr strings.Builder

	// Configure module with I/O
	config := wazero.NewModuleConfig().
		WithStdin(stdin).
		WithStdout(&stdout).
		WithStderr(&stderr).
		WithSysNanosleep().
		WithSysWalltime()

	// Instantiate the module
	mod, err := r.InstantiateWithConfig(ctx, wasmBytes, config)
	if err != nil {
		return TaskOutput{
			Status:  "failed",
			Message: fmt.Sprintf("Failed to instantiate WASM module: %v", err),
			Facts:   make(map[string]interface{}),
		}
	}
	defer mod.Close(ctx)

	// The WASM module should print JSON to stdout
	var output TaskOutput
	if err := json.Unmarshal([]byte(stdout.String()), &output); err != nil {
		return TaskOutput{
			Status:  "failed",
			Message: fmt.Sprintf("Failed to parse module output: %v\nStdout: %s\nStderr: %s", err, stdout.String(), stderr.String()),
			Facts:   make(map[string]interface{}),
		}
	}

	return output
}

func printOutput(output TaskOutput) {
	outputJSON, _ := json.MarshalIndent(output, "", "  ")
	fmt.Println(string(outputJSON))
}

// hostExec is called by WASM modules to execute shell commands
func hostExec(cmd string, args []string, timeoutSecs int) (string, string, int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeoutSecs)*time.Second)
	defer cancel()

	command := exec.CommandContext(ctx, cmd, args...)

	var stdout, stderr strings.Builder
	command.Stdout = &stdout
	command.Stderr = &stderr

	err := command.Run()
	exitCode := 0
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			exitCode = exitError.ExitCode()
		} else {
			exitCode = 1
		}
	}

	return stdout.String(), stderr.String(), exitCode, nil
}
