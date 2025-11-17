package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// Input represents the JSON input from OpenFroyo
type Input struct {
	Vars    map[string]interface{} `json:"vars"`
	Context Context                `json:"context"`
}

// Context provides execution context
type Context struct {
	Host     string `json:"host"`
	TaskName string `json:"task_name"`
}

// Output represents the JSON output to OpenFroyo
type Output struct {
	Status  string                 `json:"status"`
	Message string                 `json:"message"`
	Facts   map[string]interface{} `json:"facts"`
}

func main() {
	// Parse input
	input, err := parseInput()
	if err != nil {
		outputError(fmt.Sprintf("Failed to parse input: %v", err))
		return
	}

	// Determine operation
	operation := getString(input.Vars, "operation", "check")

	switch operation {
	case "check":
		checkKubectl(input)
	case "execute":
		executeCommand(input)
	default:
		outputError(fmt.Sprintf("Unknown operation: %s", operation))
	}
}

func parseInput() (*Input, error) {
	var input Input
	decoder := json.NewDecoder(os.Stdin)
	if err := decoder.Decode(&input); err != nil {
		return nil, err
	}
	return &input, nil
}

func checkKubectl(input *Input) {
	// Check if kubectl is installed
	cmd := exec.Command("kubectl", "version", "--client", "--short")
	output, err := cmd.CombinedOutput()

	facts := make(map[string]interface{})

	if err != nil {
		facts["installed"] = false
		facts["version"] = ""
		outputResult("failed", "kubectl is not installed or not in PATH", facts)
		return
	}

	facts["installed"] = true
	facts["version"] = strings.TrimSpace(string(output))
	outputResult("ok", "kubectl is available", facts)
}

func executeCommand(input *Input) {
	command := getString(input.Vars, "command", "")
	if command == "" {
		outputError("command parameter is required")
		return
	}

	// Execute the appropriate kubectl command
	switch command {
	// Deployment Management
	case "create_deployment":
		createDeployment(input)
	case "delete_deployment":
		deleteDeployment(input)
	case "scale_deployment":
		scaleDeployment(input)
	case "update_deployment":
		updateDeployment(input)
	case "rollout_restart":
		rolloutRestart(input)
	case "rollout_status":
		rolloutStatus(input)
	case "describe_deployment":
		describeDeployment(input)

	// Service Management
	case "create_service":
		createService(input)
	case "delete_service":
		deleteService(input)
	case "describe_service":
		describeService(input)
	case "list_services":
		listServices(input)
	case "expose_deployment":
		exposeDeployment(input)

	// Pod Management
	case "create_pod":
		createPod(input)
	case "delete_pod":
		deletePod(input)
	case "list_pods":
		listPods(input)
	case "describe_pod":
		describePod(input)
	case "get_pod_logs":
		getPodLogs(input)
	case "exec_pod":
		execPod(input)

	// ConfigMap Management
	case "create_configmap":
		createConfigMap(input)
	case "delete_configmap":
		deleteConfigMap(input)
	case "list_configmaps":
		listConfigMaps(input)

	// Secret Management
	case "create_secret":
		createSecret(input)
	case "delete_secret":
		deleteSecret(input)
	case "list_secrets":
		listSecrets(input)

	// Namespace Management
	case "create_namespace":
		createNamespace(input)
	case "delete_namespace":
		deleteNamespace(input)
	case "list_namespaces":
		listNamespaces(input)

	// Ingress Management
	case "create_ingress":
		createIngress(input)
	case "delete_ingress":
		deleteIngress(input)
	case "list_ingress":
		listIngress(input)
	case "describe_ingress":
		describeIngress(input)

	// PersistentVolumeClaim Management
	case "create_pvc":
		createPVC(input)
	case "delete_pvc":
		deletePVC(input)
	case "list_pvcs":
		listPVCs(input)
	case "describe_pvc":
		describePVC(input)

	// Additional Operations
	case "apply_manifest":
		applyManifest(input)
	case "delete_manifest":
		deleteManifest(input)
	case "get_resource":
		getResource(input)

	default:
		outputError(fmt.Sprintf("Unknown command: %s", command))
	}
}

// Deployment Management Functions

func createDeployment(input *Input) {
	name := getString(input.Vars, "deployment_name", "")
	image := getString(input.Vars, "image", "")
	namespace := getString(input.Vars, "namespace", "default")

	if name == "" || image == "" {
		outputError("deployment_name and image are required")
		return
	}

	args := []string{"create", "deployment", name, "--image=" + image, "-n", namespace}

	// Add replicas
	replicas := getInt(input.Vars, "replicas", 1)
	args = append(args, fmt.Sprintf("--replicas=%d", replicas))

	// Add port if specified
	if port := getInt(input.Vars, "port", 0); port > 0 {
		args = append(args, fmt.Sprintf("--port=%d", port))
	}

	// Execute command
	executeKubectl(args, input, fmt.Sprintf("Deployment %s created", name))
}

func deleteDeployment(input *Input) {
	name := getString(input.Vars, "deployment_name", "")
	namespace := getString(input.Vars, "namespace", "default")

	if name == "" {
		outputError("deployment_name is required")
		return
	}

	args := []string{"delete", "deployment", name, "-n", namespace}
	executeKubectl(args, input, fmt.Sprintf("Deployment %s deleted", name))
}

func scaleDeployment(input *Input) {
	name := getString(input.Vars, "deployment_name", "")
	namespace := getString(input.Vars, "namespace", "default")
	replicas := getInt(input.Vars, "replicas", 1)

	if name == "" {
		outputError("deployment_name is required")
		return
	}

	args := []string{"scale", "deployment", name, fmt.Sprintf("--replicas=%d", replicas), "-n", namespace}
	executeKubectl(args, input, fmt.Sprintf("Deployment %s scaled to %d replicas", name, replicas))
}

func updateDeployment(input *Input) {
	name := getString(input.Vars, "deployment_name", "")
	image := getString(input.Vars, "image", "")
	namespace := getString(input.Vars, "namespace", "default")

	if name == "" || image == "" {
		outputError("deployment_name and image are required")
		return
	}

	args := []string{"set", "image", "deployment/" + name, name + "=" + image, "-n", namespace}
	executeKubectl(args, input, fmt.Sprintf("Deployment %s updated with image %s", name, image))
}

func rolloutRestart(input *Input) {
	name := getString(input.Vars, "deployment_name", "")
	namespace := getString(input.Vars, "namespace", "default")

	if name == "" {
		outputError("deployment_name is required")
		return
	}

	args := []string{"rollout", "restart", "deployment/" + name, "-n", namespace}
	executeKubectl(args, input, fmt.Sprintf("Deployment %s restarted", name))
}

func rolloutStatus(input *Input) {
	name := getString(input.Vars, "deployment_name", "")
	namespace := getString(input.Vars, "namespace", "default")

	if name == "" {
		outputError("deployment_name is required")
		return
	}

	args := []string{"rollout", "status", "deployment/" + name, "-n", namespace}
	executeKubectl(args, input, fmt.Sprintf("Rollout status for deployment %s", name))
}

func describeDeployment(input *Input) {
	name := getString(input.Vars, "deployment_name", "")
	namespace := getString(input.Vars, "namespace", "default")

	if name == "" {
		outputError("deployment_name is required")
		return
	}

	args := []string{"describe", "deployment", name, "-n", namespace}
	executeKubectl(args, input, fmt.Sprintf("Description of deployment %s", name))
}

// Service Management Functions

func createService(input *Input) {
	name := getString(input.Vars, "service_name", "")
	namespace := getString(input.Vars, "namespace", "default")
	serviceType := getString(input.Vars, "service_type", "ClusterIP")

	if name == "" {
		outputError("service_name is required")
		return
	}

	// Create service using manifest
	manifest := fmt.Sprintf(`apiVersion: v1
kind: Service
metadata:
  name: %s
  namespace: %s
spec:
  type: %s
  selector:
    app: %s
  ports:
  - port: %d
    targetPort: %d
`, name, namespace, serviceType, name, getInt(input.Vars, "port", 80), getInt(input.Vars, "target_port", 80))

	executeManifest(manifest, input, fmt.Sprintf("Service %s created", name))
}

func deleteService(input *Input) {
	name := getString(input.Vars, "service_name", "")
	namespace := getString(input.Vars, "namespace", "default")

	if name == "" {
		outputError("service_name is required")
		return
	}

	args := []string{"delete", "service", name, "-n", namespace}
	executeKubectl(args, input, fmt.Sprintf("Service %s deleted", name))
}

func describeService(input *Input) {
	name := getString(input.Vars, "service_name", "")
	namespace := getString(input.Vars, "namespace", "default")

	if name == "" {
		outputError("service_name is required")
		return
	}

	args := []string{"describe", "service", name, "-n", namespace}
	executeKubectl(args, input, fmt.Sprintf("Description of service %s", name))
}

func listServices(input *Input) {
	namespace := getString(input.Vars, "namespace", "default")
	args := []string{"get", "services", "-n", namespace}
	executeKubectl(args, input, "Services listed")
}

func exposeDeployment(input *Input) {
	name := getString(input.Vars, "deployment_name", "")
	namespace := getString(input.Vars, "namespace", "default")
	serviceType := getString(input.Vars, "service_type", "ClusterIP")
	port := getInt(input.Vars, "port", 80)

	if name == "" {
		outputError("deployment_name is required")
		return
	}

	args := []string{"expose", "deployment", name, "--type=" + serviceType, fmt.Sprintf("--port=%d", port), "-n", namespace}
	executeKubectl(args, input, fmt.Sprintf("Deployment %s exposed as service", name))
}

// Pod Management Functions

func createPod(input *Input) {
	name := getString(input.Vars, "pod_name", "")
	image := getString(input.Vars, "image", "")
	namespace := getString(input.Vars, "namespace", "default")

	if name == "" || image == "" {
		outputError("pod_name and image are required")
		return
	}

	args := []string{"run", name, "--image=" + image, "-n", namespace}
	executeKubectl(args, input, fmt.Sprintf("Pod %s created", name))
}

func deletePod(input *Input) {
	name := getString(input.Vars, "pod_name", "")
	namespace := getString(input.Vars, "namespace", "default")

	if name == "" {
		outputError("pod_name is required")
		return
	}

	args := []string{"delete", "pod", name, "-n", namespace}
	executeKubectl(args, input, fmt.Sprintf("Pod %s deleted", name))
}

func listPods(input *Input) {
	namespace := getString(input.Vars, "namespace", "default")
	args := []string{"get", "pods", "-n", namespace}

	if getBool(input.Vars, "all_namespaces", false) {
		args = []string{"get", "pods", "--all-namespaces"}
	}

	executeKubectl(args, input, "Pods listed")
}

func describePod(input *Input) {
	name := getString(input.Vars, "pod_name", "")
	namespace := getString(input.Vars, "namespace", "default")

	if name == "" {
		outputError("pod_name is required")
		return
	}

	args := []string{"describe", "pod", name, "-n", namespace}
	executeKubectl(args, input, fmt.Sprintf("Description of pod %s", name))
}

func getPodLogs(input *Input) {
	name := getString(input.Vars, "pod_name", "")
	namespace := getString(input.Vars, "namespace", "default")

	if name == "" {
		outputError("pod_name is required")
		return
	}

	args := []string{"logs", name, "-n", namespace}

	// Add container name if specified
	if container := getString(input.Vars, "container_name", ""); container != "" {
		args = append(args, "-c", container)
	}

	// Add tail lines
	if tail := getInt(input.Vars, "log_tail_lines", 0); tail > 0 {
		args = append(args, fmt.Sprintf("--tail=%d", tail))
	}

	// Add follow flag
	if getBool(input.Vars, "log_follow", false) {
		args = append(args, "-f")
	}

	// Add timestamps
	if getBool(input.Vars, "log_timestamps", false) {
		args = append(args, "--timestamps")
	}

	// Add previous flag
	if getBool(input.Vars, "log_previous", false) {
		args = append(args, "--previous")
	}

	executeKubectl(args, input, fmt.Sprintf("Logs for pod %s", name))
}

func execPod(input *Input) {
	name := getString(input.Vars, "pod_name", "")
	namespace := getString(input.Vars, "namespace", "default")

	if name == "" {
		outputError("pod_name is required")
		return
	}

	args := []string{"exec", name, "-n", namespace}

	// Add container name if specified
	if container := getString(input.Vars, "container_name", ""); container != "" {
		args = append(args, "-c", container)
	}

	// Add stdin flag
	if getBool(input.Vars, "exec_stdin", false) {
		args = append(args, "-i")
	}

	// Add tty flag
	if getBool(input.Vars, "exec_tty", false) {
		args = append(args, "-t")
	}

	args = append(args, "--")

	// Add command
	if cmdList, ok := input.Vars["exec_command"].([]interface{}); ok {
		for _, cmd := range cmdList {
			args = append(args, fmt.Sprintf("%v", cmd))
		}
	}

	executeKubectl(args, input, fmt.Sprintf("Executed command in pod %s", name))
}

// ConfigMap Management Functions

func createConfigMap(input *Input) {
	name := getString(input.Vars, "configmap_name", "")
	namespace := getString(input.Vars, "namespace", "default")

	if name == "" {
		outputError("configmap_name is required")
		return
	}

	args := []string{"create", "configmap", name, "-n", namespace}

	// Add data from literal values
	if data, ok := input.Vars["configmap_data"].(map[string]interface{}); ok {
		for key, value := range data {
			args = append(args, fmt.Sprintf("--from-literal=%s=%v", key, value))
		}
	}

	// Add data from file
	if file := getString(input.Vars, "configmap_from_file", ""); file != "" {
		args = append(args, "--from-file="+file)
	}

	executeKubectl(args, input, fmt.Sprintf("ConfigMap %s created", name))
}

func deleteConfigMap(input *Input) {
	name := getString(input.Vars, "configmap_name", "")
	namespace := getString(input.Vars, "namespace", "default")

	if name == "" {
		outputError("configmap_name is required")
		return
	}

	args := []string{"delete", "configmap", name, "-n", namespace}
	executeKubectl(args, input, fmt.Sprintf("ConfigMap %s deleted", name))
}

func listConfigMaps(input *Input) {
	namespace := getString(input.Vars, "namespace", "default")
	args := []string{"get", "configmaps", "-n", namespace}
	executeKubectl(args, input, "ConfigMaps listed")
}

// Secret Management Functions

func createSecret(input *Input) {
	name := getString(input.Vars, "secret_name", "")
	namespace := getString(input.Vars, "namespace", "default")
	secretType := getString(input.Vars, "secret_type", "generic")

	if name == "" {
		outputError("secret_name is required")
		return
	}

	args := []string{"create", "secret", secretType, name, "-n", namespace}

	// Add data from literal values
	if data, ok := input.Vars["secret_data"].(map[string]interface{}); ok {
		for key, value := range data {
			args = append(args, fmt.Sprintf("--from-literal=%s=%v", key, value))
		}
	}

	// Add data from file
	if file := getString(input.Vars, "secret_from_file", ""); file != "" {
		args = append(args, "--from-file="+file)
	}

	executeKubectl(args, input, fmt.Sprintf("Secret %s created", name))
}

func deleteSecret(input *Input) {
	name := getString(input.Vars, "secret_name", "")
	namespace := getString(input.Vars, "namespace", "default")

	if name == "" {
		outputError("secret_name is required")
		return
	}

	args := []string{"delete", "secret", name, "-n", namespace}
	executeKubectl(args, input, fmt.Sprintf("Secret %s deleted", name))
}

func listSecrets(input *Input) {
	namespace := getString(input.Vars, "namespace", "default")
	args := []string{"get", "secrets", "-n", namespace}
	executeKubectl(args, input, "Secrets listed")
}

// Namespace Management Functions

func createNamespace(input *Input) {
	name := getString(input.Vars, "namespace_name", "")

	if name == "" {
		outputError("namespace_name is required")
		return
	}

	args := []string{"create", "namespace", name}
	executeKubectl(args, input, fmt.Sprintf("Namespace %s created", name))
}

func deleteNamespace(input *Input) {
	name := getString(input.Vars, "namespace_name", "")

	if name == "" {
		outputError("namespace_name is required")
		return
	}

	args := []string{"delete", "namespace", name}
	executeKubectl(args, input, fmt.Sprintf("Namespace %s deleted", name))
}

func listNamespaces(input *Input) {
	args := []string{"get", "namespaces"}
	executeKubectl(args, input, "Namespaces listed")
}

// Ingress Management Functions

func createIngress(input *Input) {
	name := getString(input.Vars, "ingress_name", "")
	namespace := getString(input.Vars, "namespace", "default")
	host := getString(input.Vars, "ingress_host", "")
	serviceName := getString(input.Vars, "ingress_service_name", "")
	servicePort := getInt(input.Vars, "ingress_service_port", 80)

	if name == "" || host == "" || serviceName == "" {
		outputError("ingress_name, ingress_host, and ingress_service_name are required")
		return
	}

	path := getString(input.Vars, "ingress_path", "/")
	rule := fmt.Sprintf("%s%s=%s:%d", host, path, serviceName, servicePort)

	args := []string{"create", "ingress", name, "--rule=" + rule, "-n", namespace}
	executeKubectl(args, input, fmt.Sprintf("Ingress %s created", name))
}

func deleteIngress(input *Input) {
	name := getString(input.Vars, "ingress_name", "")
	namespace := getString(input.Vars, "namespace", "default")

	if name == "" {
		outputError("ingress_name is required")
		return
	}

	args := []string{"delete", "ingress", name, "-n", namespace}
	executeKubectl(args, input, fmt.Sprintf("Ingress %s deleted", name))
}

func listIngress(input *Input) {
	namespace := getString(input.Vars, "namespace", "default")
	args := []string{"get", "ingress", "-n", namespace}
	executeKubectl(args, input, "Ingress resources listed")
}

func describeIngress(input *Input) {
	name := getString(input.Vars, "ingress_name", "")
	namespace := getString(input.Vars, "namespace", "default")

	if name == "" {
		outputError("ingress_name is required")
		return
	}

	args := []string{"describe", "ingress", name, "-n", namespace}
	executeKubectl(args, input, fmt.Sprintf("Description of ingress %s", name))
}

// PVC Management Functions

func createPVC(input *Input) {
	name := getString(input.Vars, "pvc_name", "")
	namespace := getString(input.Vars, "namespace", "default")
	size := getString(input.Vars, "storage_size", "10Gi")
	accessMode := getString(input.Vars, "access_mode", "ReadWriteOnce")

	if name == "" {
		outputError("pvc_name is required")
		return
	}

	manifest := fmt.Sprintf(`apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: %s
  namespace: %s
spec:
  accessModes:
  - %s
  resources:
    requests:
      storage: %s
`, name, namespace, accessMode, size)

	// Add storage class if specified
	if storageClass := getString(input.Vars, "storage_class", ""); storageClass != "" {
		manifest += fmt.Sprintf("  storageClassName: %s\n", storageClass)
	}

	executeManifest(manifest, input, fmt.Sprintf("PVC %s created", name))
}

func deletePVC(input *Input) {
	name := getString(input.Vars, "pvc_name", "")
	namespace := getString(input.Vars, "namespace", "default")

	if name == "" {
		outputError("pvc_name is required")
		return
	}

	args := []string{"delete", "pvc", name, "-n", namespace}
	executeKubectl(args, input, fmt.Sprintf("PVC %s deleted", name))
}

func listPVCs(input *Input) {
	namespace := getString(input.Vars, "namespace", "default")
	args := []string{"get", "pvc", "-n", namespace}
	executeKubectl(args, input, "PVCs listed")
}

func describePVC(input *Input) {
	name := getString(input.Vars, "pvc_name", "")
	namespace := getString(input.Vars, "namespace", "default")

	if name == "" {
		outputError("pvc_name is required")
		return
	}

	args := []string{"describe", "pvc", name, "-n", namespace}
	executeKubectl(args, input, fmt.Sprintf("Description of PVC %s", name))
}

// Additional Operations

func applyManifest(input *Input) {
	manifestFile := getString(input.Vars, "manifest_file", "")
	manifestContent := getString(input.Vars, "manifest_content", "")

	if manifestFile != "" {
		args := []string{"apply", "-f", manifestFile}
		executeKubectl(args, input, "Manifest applied from file")
	} else if manifestContent != "" {
		executeManifest(manifestContent, input, "Manifest applied from content")
	} else {
		outputError("Either manifest_file or manifest_content is required")
	}
}

func deleteManifest(input *Input) {
	manifestFile := getString(input.Vars, "manifest_file", "")

	if manifestFile == "" {
		outputError("manifest_file is required")
		return
	}

	args := []string{"delete", "-f", manifestFile}
	executeKubectl(args, input, "Resources deleted from manifest")
}

func getResource(input *Input) {
	resourceType := getString(input.Vars, "resource_type", "")
	resourceName := getString(input.Vars, "resource_name", "")
	namespace := getString(input.Vars, "namespace", "default")

	if resourceType == "" {
		outputError("resource_type is required")
		return
	}

	args := []string{"get", resourceType}

	if resourceName != "" {
		args = append(args, resourceName)
	}

	args = append(args, "-n", namespace)

	outputFormat := getString(input.Vars, "output_format", "wide")
	if outputFormat != "" {
		args = append(args, "-o", outputFormat)
	}

	executeKubectl(args, input, fmt.Sprintf("Resource %s retrieved", resourceType))
}

// Helper Functions

func executeKubectl(args []string, input *Input, successMsg string) {
	// Add kubeconfig if specified
	if kubeconfig := getString(input.Vars, "kubeconfig", ""); kubeconfig != "" {
		args = append([]string{"--kubeconfig=" + kubeconfig}, args...)
	}

	// Add context if specified
	if context := getString(input.Vars, "context", ""); context != "" {
		args = append([]string{"--context=" + context}, args...)
	}

	cmd := exec.Command("kubectl", args...)
	output, err := cmd.CombinedOutput()

	facts := make(map[string]interface{})
	facts["output"] = string(output)
	facts["command"] = "kubectl " + strings.Join(args, " ")

	if err != nil {
		// Check if it's a "no changes" or "already exists" error (not critical)
		if strings.Contains(string(output), "unchanged") ||
		   strings.Contains(string(output), "already exists") {
			outputResult("ok", string(output), facts)
		} else {
			facts["error"] = err.Error()
			outputResult("failed", fmt.Sprintf("kubectl command failed: %v", err), facts)
		}
		return
	}

	// Determine if changes were made
	status := "ok"
	if strings.Contains(string(output), "created") ||
	   strings.Contains(string(output), "deleted") ||
	   strings.Contains(string(output), "scaled") ||
	   strings.Contains(string(output), "updated") {
		status = "changed"
	}

	outputResult(status, successMsg, facts)
}

func executeManifest(manifest string, input *Input, successMsg string) {
	// Create temporary file for manifest
	tmpFile, err := os.CreateTemp("", "k8s-manifest-*.yaml")
	if err != nil {
		outputError(fmt.Sprintf("Failed to create temporary file: %v", err))
		return
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(manifest); err != nil {
		outputError(fmt.Sprintf("Failed to write manifest: %v", err))
		return
	}
	tmpFile.Close()

	args := []string{"apply", "-f", tmpFile.Name()}
	executeKubectl(args, input, successMsg)
}

func getString(vars map[string]interface{}, key string, defaultValue string) string {
	if val, ok := vars[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return defaultValue
}

func getInt(vars map[string]interface{}, key string, defaultValue int) int {
	if val, ok := vars[key]; ok {
		switch v := val.(type) {
		case int:
			return v
		case float64:
			return int(v)
		case string:
			var i int
			fmt.Sscanf(v, "%d", &i)
			return i
		}
	}
	return defaultValue
}

func getBool(vars map[string]interface{}, key string, defaultValue bool) bool {
	if val, ok := vars[key]; ok {
		if b, ok := val.(bool); ok {
			return b
		}
	}
	return defaultValue
}

func outputResult(status, message string, facts map[string]interface{}) {
	output := Output{
		Status:  status,
		Message: message,
		Facts:   facts,
	}

	encoder := json.NewEncoder(os.Stdout)
	encoder.Encode(output)
}

func outputError(message string) {
	output := Output{
		Status:  "failed",
		Message: message,
		Facts:   make(map[string]interface{}),
	}

	encoder := json.NewEncoder(os.Stdout)
	encoder.Encode(output)
}
