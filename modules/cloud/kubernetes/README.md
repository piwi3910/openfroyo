# Kubernetes Module

Comprehensive Kubernetes cluster management module for OpenFroyo. This module provides kubectl-based operations for managing Kubernetes resources across your cluster.

## Overview

The Kubernetes module enables you to:
- Deploy and manage applications using Deployments, StatefulSets, and DaemonSets
- Expose services with various service types (ClusterIP, NodePort, LoadBalancer)
- Manage configuration with ConfigMaps and Secrets
- Control access with Namespaces and RBAC
- Configure ingress routing for external access
- Provision storage with PersistentVolumeClaims
- Monitor and debug with logs, exec, and describe operations

## Requirements

- `kubectl` must be installed on target hosts
- Valid kubeconfig file or in-cluster service account credentials
- Appropriate RBAC permissions for the operations you want to perform

## Installation

The module is located at `modules/cloud/kubernetes/` and includes:
- `module.ofy.yml` - Module definition
- `defaults.ofy.yml` - Default variable values
- `wasm/kubernetes.wasm` - Compiled WebAssembly module
- `wasm/main.go` - Source code for the WASM module

## Quick Start

### Create a Simple Deployment

```yaml
- name: Deploy nginx application
  module: cloud/kubernetes
  vars:
    command: create_deployment
    deployment_name: nginx-app
    image: nginx:latest
    replicas: 3
    port: 80
    namespace: production
```

### Expose a Service

```yaml
- name: Expose nginx service
  module: cloud/kubernetes
  vars:
    command: expose_deployment
    deployment_name: nginx-app
    service_type: LoadBalancer
    port: 80
    namespace: production
```

### Create a ConfigMap

```yaml
- name: Create application configuration
  module: cloud/kubernetes
  vars:
    command: create_configmap
    configmap_name: app-config
    namespace: production
    configmap_data:
      database_host: db.example.com
      cache_ttl: "3600"
      feature_flags: enabled
```

### Create a Secret

```yaml
- name: Create database credentials
  module: cloud/kubernetes
  vars:
    command: create_secret
    secret_name: db-creds
    secret_type: generic
    namespace: production
    secret_data:
      username: dbuser
      password: securepassword123
```

## Core Concepts

### Namespaces

Kubernetes namespaces provide logical isolation for resources. By default, the module operates in the `default` namespace, but you can specify any namespace:

```yaml
vars:
  namespace: production
```

Create a new namespace:

```yaml
- name: Create production namespace
  module: cloud/kubernetes
  vars:
    command: create_namespace
    namespace_name: production
```

### Deployments

Deployments manage the lifecycle of your application pods:

```yaml
- name: Deploy web application
  module: cloud/kubernetes
  vars:
    command: create_deployment
    deployment_name: webapp
    image: myapp:v1.2.3
    replicas: 5
    port: 8080
    namespace: production
```

Update a deployment with a new image:

```yaml
- name: Update application
  module: cloud/kubernetes
  vars:
    command: update_deployment
    deployment_name: webapp
    image: myapp:v1.2.4
    namespace: production
```

Scale a deployment:

```yaml
- name: Scale application
  module: cloud/kubernetes
  vars:
    command: scale_deployment
    deployment_name: webapp
    replicas: 10
    namespace: production
```

### Services

Services provide stable networking endpoints for your pods:

```yaml
- name: Create LoadBalancer service
  module: cloud/kubernetes
  vars:
    command: create_service
    service_name: webapp-lb
    service_type: LoadBalancer
    port: 80
    target_port: 8080
    namespace: production
```

Service types:
- `ClusterIP` - Internal cluster IP (default)
- `NodePort` - Exposes service on each node's IP at a static port
- `LoadBalancer` - Creates an external load balancer (cloud provider)
- `ExternalName` - Maps service to a DNS name

### ConfigMaps and Secrets

ConfigMaps store non-sensitive configuration:

```yaml
- name: Create app config
  module: cloud/kubernetes
  vars:
    command: create_configmap
    configmap_name: app-settings
    namespace: production
    configmap_data:
      log_level: info
      max_connections: "100"
      enable_metrics: "true"
```

Secrets store sensitive data (base64 encoded):

```yaml
- name: Create TLS certificate
  module: cloud/kubernetes
  vars:
    command: create_secret
    secret_name: tls-cert
    secret_type: kubernetes.io/tls
    namespace: production
    secret_from_file: /path/to/cert.pem
```

### Ingress

Ingress provides HTTP/HTTPS routing to services:

```yaml
- name: Create ingress rule
  module: cloud/kubernetes
  vars:
    command: create_ingress
    ingress_name: webapp-ingress
    ingress_host: webapp.example.com
    ingress_path: /
    ingress_service_name: webapp
    ingress_service_port: 80
    namespace: production
```

### Persistent Storage

PersistentVolumeClaims request storage resources:

```yaml
- name: Create storage for database
  module: cloud/kubernetes
  vars:
    command: create_pvc
    pvc_name: db-storage
    storage_size: 50Gi
    storage_class: fast-ssd
    access_mode: ReadWriteOnce
    namespace: production
```

Access modes:
- `ReadWriteOnce` - Single node read/write
- `ReadOnlyMany` - Multiple nodes read-only
- `ReadWriteMany` - Multiple nodes read/write

## Configuration

### Connection Settings

```yaml
vars:
  kubeconfig: /path/to/kubeconfig  # Optional, uses default if not specified
  context: production-cluster       # Optional, uses current context if not specified
  namespace: default                # Target namespace
```

### Common Parameters

All commands support these common parameters:

```yaml
vars:
  namespace: default           # Target namespace
  dry_run: false              # Perform dry-run without making changes
  wait: true                  # Wait for operation to complete
  timeout: 5m                 # Operation timeout
  output_format: wide         # Output format: wide, json, yaml, name
```

## Operations Reference

### Deployment Management (7 operations)

1. **create_deployment** - Create a new deployment
2. **delete_deployment** - Delete a deployment
3. **scale_deployment** - Scale deployment replicas
4. **update_deployment** - Update deployment image
5. **rollout_restart** - Restart deployment rollout
6. **rollout_status** - Check rollout status
7. **describe_deployment** - Get deployment details

### Service Management (5 operations)

1. **create_service** - Create a new service
2. **delete_service** - Delete a service
3. **describe_service** - Get service details
4. **list_services** - List all services
5. **expose_deployment** - Expose deployment as service

### Pod Management (6 operations)

1. **create_pod** - Create a standalone pod
2. **delete_pod** - Delete a pod
3. **list_pods** - List all pods
4. **describe_pod** - Get pod details
5. **get_pod_logs** - Retrieve pod logs
6. **exec_pod** - Execute command in pod

### ConfigMap Management (3 operations)

1. **create_configmap** - Create a ConfigMap
2. **delete_configmap** - Delete a ConfigMap
3. **list_configmaps** - List all ConfigMaps

### Secret Management (3 operations)

1. **create_secret** - Create a Secret
2. **delete_secret** - Delete a Secret
3. **list_secrets** - List all Secrets

### Namespace Management (3 operations)

1. **create_namespace** - Create a namespace
2. **delete_namespace** - Delete a namespace
3. **list_namespaces** - List all namespaces

### Ingress Management (4 operations)

1. **create_ingress** - Create an ingress rule
2. **delete_ingress** - Delete an ingress
3. **list_ingress** - List all ingress resources
4. **describe_ingress** - Get ingress details

### PersistentVolumeClaim Management (4 operations)

1. **create_pvc** - Create a PVC
2. **delete_pvc** - Delete a PVC
3. **list_pvcs** - List all PVCs
4. **describe_pvc** - Get PVC details

### Additional Operations (3 operations)

1. **apply_manifest** - Apply YAML manifest
2. **delete_manifest** - Delete resources from manifest
3. **get_resource** - Get any resource type

**Total: 38 operations**

## Advanced Usage

### Working with Manifests

Apply a manifest file:

```yaml
- name: Apply application manifest
  module: cloud/kubernetes
  vars:
    command: apply_manifest
    manifest_file: /path/to/deployment.yaml
    namespace: production
```

Apply inline manifest content:

```yaml
- name: Create custom resource
  module: cloud/kubernetes
  vars:
    command: apply_manifest
    namespace: production
    manifest_content: |
      apiVersion: v1
      kind: Service
      metadata:
        name: my-service
      spec:
        selector:
          app: myapp
        ports:
        - protocol: TCP
          port: 80
          targetPort: 8080
```

### Getting Pod Logs

Retrieve recent logs:

```yaml
- name: Get application logs
  module: cloud/kubernetes
  vars:
    command: get_pod_logs
    pod_name: webapp-7d8f9c-abcde
    namespace: production
    log_tail_lines: 100
    log_timestamps: true
```

Follow logs in real-time:

```yaml
- name: Follow application logs
  module: cloud/kubernetes
  vars:
    command: get_pod_logs
    pod_name: webapp-7d8f9c-abcde
    container_name: app
    namespace: production
    log_follow: true
```

### Executing Commands in Pods

Run a command in a pod:

```yaml
- name: Run database migration
  module: cloud/kubernetes
  vars:
    command: exec_pod
    pod_name: webapp-7d8f9c-abcde
    namespace: production
    exec_command:
      - /bin/sh
      - -c
      - python manage.py migrate
```

Interactive shell:

```yaml
- name: Open shell in pod
  module: cloud/kubernetes
  vars:
    command: exec_pod
    pod_name: webapp-7d8f9c-abcde
    namespace: production
    exec_stdin: true
    exec_tty: true
    exec_command:
      - /bin/bash
```

### Query Resources

Get resources with custom output:

```yaml
- name: Get pods as JSON
  module: cloud/kubernetes
  vars:
    command: get_resource
    resource_type: pods
    namespace: production
    output_format: json
```

Get specific resource:

```yaml
- name: Get deployment details
  module: cloud/kubernetes
  vars:
    command: get_resource
    resource_type: deployment
    resource_name: webapp
    namespace: production
    output_format: yaml
```

## Best Practices

### 1. Use Namespaces for Isolation

Always specify namespaces to avoid accidental operations on the wrong environment:

```yaml
defaults:
  namespace: production
```

### 2. Label Everything

Use labels for organization and selection:

```yaml
vars:
  labels:
    app: webapp
    environment: production
    version: v1.2.3
    team: platform
```

### 3. Resource Limits

Always set resource limits for production workloads (use manifests for complex configurations):

```yaml
manifest_content: |
  resources:
    limits:
      cpu: "1000m"
      memory: "1Gi"
    requests:
      cpu: "500m"
      memory: "512Mi"
```

### 4. Health Checks

Configure liveness and readiness probes (via manifests):

```yaml
manifest_content: |
  livenessProbe:
    httpGet:
      path: /health
      port: 8080
    initialDelaySeconds: 30
    periodSeconds: 10
  readinessProbe:
    httpGet:
      path: /ready
      port: 8080
    initialDelaySeconds: 5
    periodSeconds: 5
```

### 5. ConfigMaps and Secrets

- Use ConfigMaps for non-sensitive configuration
- Use Secrets for sensitive data (passwords, tokens, keys)
- Never commit secrets to version control
- Consider external secret management (Vault, AWS Secrets Manager)

### 6. Rolling Updates

Configure rolling update strategy for zero-downtime deployments:

```yaml
manifest_content: |
  strategy:
    type: RollingUpdate
    rollingUpdate:
      maxSurge: 1
      maxUnavailable: 0
```

### 7. Network Policies

Implement network policies for security:

```yaml
- name: Create network policy
  module: cloud/kubernetes
  vars:
    command: apply_manifest
    manifest_content: |
      apiVersion: networking.k8s.io/v1
      kind: NetworkPolicy
      metadata:
        name: webapp-netpol
      spec:
        podSelector:
          matchLabels:
            app: webapp
        policyTypes:
        - Ingress
        - Egress
        ingress:
        - from:
          - podSelector:
              matchLabels:
                role: frontend
```

## Security Considerations

### 1. RBAC Permissions

Ensure your kubeconfig has appropriate permissions:

- Minimum: Read-only access for monitoring
- Standard: Create/update/delete for deployments and services
- Admin: Full cluster access (use carefully)

### 2. Service Accounts

Use service accounts for applications:

```yaml
- name: Create service account
  module: cloud/kubernetes
  vars:
    command: apply_manifest
    manifest_content: |
      apiVersion: v1
      kind: ServiceAccount
      metadata:
        name: webapp-sa
        namespace: production
```

### 3. Secrets Management

- Rotate secrets regularly
- Use encrypted secrets at rest
- Consider external secret stores
- Limit secret access with RBAC

### 4. Image Security

- Use specific image tags, not `latest`
- Scan images for vulnerabilities
- Use private registries for production
- Implement image pull policies

### 5. Pod Security

- Run containers as non-root
- Use read-only root filesystems
- Drop unnecessary capabilities
- Implement pod security policies

## Troubleshooting

### kubectl Not Found

**Error:** `kubectl is not installed or not in PATH`

**Solution:** Install kubectl on the target host:

```bash
# Linux
curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl"
chmod +x kubectl
sudo mv kubectl /usr/local/bin/

# macOS
brew install kubectl
```

### Connection Refused

**Error:** `The connection to the server was refused`

**Solution:** Check kubeconfig and cluster availability:

```yaml
vars:
  kubeconfig: /home/user/.kube/config
  context: production-cluster
```

### Permission Denied

**Error:** `User cannot create resource in API group`

**Solution:** Check RBAC permissions:

```bash
kubectl auth can-i create deployments --namespace production
```

### Pod Not Starting

**Issue:** Pod stays in Pending or CrashLoopBackOff

**Debugging:**

```yaml
- name: Describe pod
  module: cloud/kubernetes
  vars:
    command: describe_pod
    pod_name: webapp-xxx
    namespace: production

- name: Get pod logs
  module: cloud/kubernetes
  vars:
    command: get_pod_logs
    pod_name: webapp-xxx
    namespace: production
    log_previous: true
```

### Resource Quota Exceeded

**Error:** `exceeded quota`

**Solution:** Check namespace resource quotas:

```yaml
- name: Check quotas
  module: cloud/kubernetes
  vars:
    command: get_resource
    resource_type: resourcequota
    namespace: production
```

## Examples

See `test.ofy` for comprehensive examples including:
- Creating namespaces
- Deploying applications
- Exposing services
- Managing configuration
- Storage provisioning
- Ingress setup

## Reference Documentation

- [Kubernetes Official Documentation](https://kubernetes.io/docs/)
- [kubectl Command Reference](https://kubernetes.io/docs/reference/kubectl/)
- [Kubernetes API Reference](https://kubernetes.io/docs/reference/kubernetes-api/)
- [Best Practices](https://kubernetes.io/docs/concepts/configuration/overview/)

## Module Development

### Building the WASM Module

```bash
cd modules/cloud/kubernetes/wasm
make build
```

### Testing

```bash
# Run the test stack
froyo apply modules/cloud/kubernetes/test.ofy

# Validate with kubectl
kubectl get all -n froyo-test
```

### Customization

To add new operations:

1. Add the operation to `wasm/main.go`
2. Implement the function following existing patterns
3. Rebuild the WASM module with `make build`
4. Update documentation

## License

Part of the OpenFroyo project.
