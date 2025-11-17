# Kubernetes Module - Command Reference

Complete reference for all 38+ operations supported by the Kubernetes module.

## Table of Contents

- [Deployment Management](#deployment-management)
- [Service Management](#service-management)
- [Pod Management](#pod-management)
- [ConfigMap Management](#configmap-management)
- [Secret Management](#secret-management)
- [Namespace Management](#namespace-management)
- [Ingress Management](#ingress-management)
- [PersistentVolumeClaim Management](#persistentvolumeclaim-management)
- [Additional Operations](#additional-operations)

---

## Deployment Management

### create_deployment

Create a new deployment with specified image and replicas.

**Required Parameters:**
- `deployment_name` - Name of the deployment
- `image` - Container image to deploy

**Optional Parameters:**
- `replicas` - Number of pod replicas (default: 1)
- `port` - Container port to expose
- `namespace` - Target namespace (default: "default")
- `labels` - Labels to apply to the deployment

**kubectl Equivalent:**
```bash
kubectl create deployment <name> --image=<image> --replicas=<count> --port=<port> -n <namespace>
```

**Example:**
```yaml
- name: Deploy web application
  module: cloud/kubernetes
  vars:
    command: create_deployment
    deployment_name: webapp
    image: nginx:1.21
    replicas: 3
    port: 80
    namespace: production
    labels:
      app: webapp
      tier: frontend
```

---

### delete_deployment

Delete an existing deployment.

**Required Parameters:**
- `deployment_name` - Name of the deployment to delete

**Optional Parameters:**
- `namespace` - Target namespace (default: "default")

**kubectl Equivalent:**
```bash
kubectl delete deployment <name> -n <namespace>
```

**Example:**
```yaml
- name: Remove old deployment
  module: cloud/kubernetes
  vars:
    command: delete_deployment
    deployment_name: old-webapp
    namespace: production
```

---

### scale_deployment

Scale a deployment to a specific number of replicas.

**Required Parameters:**
- `deployment_name` - Name of the deployment
- `replicas` - Desired number of replicas

**Optional Parameters:**
- `namespace` - Target namespace (default: "default")

**kubectl Equivalent:**
```bash
kubectl scale deployment <name> --replicas=<count> -n <namespace>
```

**Example:**
```yaml
- name: Scale up for peak traffic
  module: cloud/kubernetes
  vars:
    command: scale_deployment
    deployment_name: webapp
    replicas: 10
    namespace: production
```

---

### update_deployment

Update the container image of a deployment.

**Required Parameters:**
- `deployment_name` - Name of the deployment
- `image` - New container image

**Optional Parameters:**
- `namespace` - Target namespace (default: "default")

**kubectl Equivalent:**
```bash
kubectl set image deployment/<name> <name>=<image> -n <namespace>
```

**Example:**
```yaml
- name: Deploy new version
  module: cloud/kubernetes
  vars:
    command: update_deployment
    deployment_name: webapp
    image: webapp:v2.0.0
    namespace: production
```

---

### rollout_restart

Restart all pods in a deployment (rolling restart).

**Required Parameters:**
- `deployment_name` - Name of the deployment

**Optional Parameters:**
- `namespace` - Target namespace (default: "default")

**kubectl Equivalent:**
```bash
kubectl rollout restart deployment/<name> -n <namespace>
```

**Example:**
```yaml
- name: Restart application
  module: cloud/kubernetes
  vars:
    command: rollout_restart
    deployment_name: webapp
    namespace: production
```

---

### rollout_status

Check the status of a deployment rollout.

**Required Parameters:**
- `deployment_name` - Name of the deployment

**Optional Parameters:**
- `namespace` - Target namespace (default: "default")

**kubectl Equivalent:**
```bash
kubectl rollout status deployment/<name> -n <namespace>
```

**Example:**
```yaml
- name: Check deployment status
  module: cloud/kubernetes
  vars:
    command: rollout_status
    deployment_name: webapp
    namespace: production
```

---

### describe_deployment

Get detailed information about a deployment.

**Required Parameters:**
- `deployment_name` - Name of the deployment

**Optional Parameters:**
- `namespace` - Target namespace (default: "default")

**kubectl Equivalent:**
```bash
kubectl describe deployment <name> -n <namespace>
```

**Example:**
```yaml
- name: Get deployment details
  module: cloud/kubernetes
  vars:
    command: describe_deployment
    deployment_name: webapp
    namespace: production
```

---

## Service Management

### create_service

Create a new Kubernetes service.

**Required Parameters:**
- `service_name` - Name of the service

**Optional Parameters:**
- `service_type` - Type of service (default: "ClusterIP")
  - ClusterIP
  - NodePort
  - LoadBalancer
  - ExternalName
- `port` - Service port (default: 80)
- `target_port` - Target port on pods (default: 80)
- `namespace` - Target namespace (default: "default")

**kubectl Equivalent:**
```bash
# Created via manifest
```

**Example:**
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

---

### delete_service

Delete a Kubernetes service.

**Required Parameters:**
- `service_name` - Name of the service

**Optional Parameters:**
- `namespace` - Target namespace (default: "default")

**kubectl Equivalent:**
```bash
kubectl delete service <name> -n <namespace>
```

**Example:**
```yaml
- name: Remove service
  module: cloud/kubernetes
  vars:
    command: delete_service
    service_name: webapp-lb
    namespace: production
```

---

### describe_service

Get detailed information about a service.

**Required Parameters:**
- `service_name` - Name of the service

**Optional Parameters:**
- `namespace` - Target namespace (default: "default")

**kubectl Equivalent:**
```bash
kubectl describe service <name> -n <namespace>
```

**Example:**
```yaml
- name: Get service details
  module: cloud/kubernetes
  vars:
    command: describe_service
    service_name: webapp-lb
    namespace: production
```

---

### list_services

List all services in a namespace.

**Optional Parameters:**
- `namespace` - Target namespace (default: "default")

**kubectl Equivalent:**
```bash
kubectl get services -n <namespace>
```

**Example:**
```yaml
- name: List all services
  module: cloud/kubernetes
  vars:
    command: list_services
    namespace: production
```

---

### expose_deployment

Expose a deployment as a service.

**Required Parameters:**
- `deployment_name` - Name of the deployment to expose

**Optional Parameters:**
- `service_type` - Type of service (default: "ClusterIP")
- `port` - Service port (default: 80)
- `namespace` - Target namespace (default: "default")

**kubectl Equivalent:**
```bash
kubectl expose deployment <name> --type=<type> --port=<port> -n <namespace>
```

**Example:**
```yaml
- name: Expose deployment
  module: cloud/kubernetes
  vars:
    command: expose_deployment
    deployment_name: webapp
    service_type: LoadBalancer
    port: 80
    namespace: production
```

---

## Pod Management

### create_pod

Create a standalone pod (not part of a deployment).

**Required Parameters:**
- `pod_name` - Name of the pod
- `image` - Container image

**Optional Parameters:**
- `namespace` - Target namespace (default: "default")

**kubectl Equivalent:**
```bash
kubectl run <name> --image=<image> -n <namespace>
```

**Example:**
```yaml
- name: Create debug pod
  module: cloud/kubernetes
  vars:
    command: create_pod
    pod_name: debug-pod
    image: busybox:latest
    namespace: production
```

---

### delete_pod

Delete a pod.

**Required Parameters:**
- `pod_name` - Name of the pod

**Optional Parameters:**
- `namespace` - Target namespace (default: "default")

**kubectl Equivalent:**
```bash
kubectl delete pod <name> -n <namespace>
```

**Example:**
```yaml
- name: Remove pod
  module: cloud/kubernetes
  vars:
    command: delete_pod
    pod_name: debug-pod
    namespace: production
```

---

### list_pods

List all pods in a namespace.

**Optional Parameters:**
- `namespace` - Target namespace (default: "default")
- `all_namespaces` - List pods across all namespaces (default: false)

**kubectl Equivalent:**
```bash
kubectl get pods -n <namespace>
# or
kubectl get pods --all-namespaces
```

**Example:**
```yaml
- name: List all pods
  module: cloud/kubernetes
  vars:
    command: list_pods
    namespace: production
```

---

### describe_pod

Get detailed information about a pod.

**Required Parameters:**
- `pod_name` - Name of the pod

**Optional Parameters:**
- `namespace` - Target namespace (default: "default")

**kubectl Equivalent:**
```bash
kubectl describe pod <name> -n <namespace>
```

**Example:**
```yaml
- name: Get pod details
  module: cloud/kubernetes
  vars:
    command: describe_pod
    pod_name: webapp-abc123
    namespace: production
```

---

### get_pod_logs

Retrieve logs from a pod.

**Required Parameters:**
- `pod_name` - Name of the pod

**Optional Parameters:**
- `container_name` - Specific container name (for multi-container pods)
- `log_tail_lines` - Number of lines to show (default: 100)
- `log_follow` - Follow log output (default: false)
- `log_timestamps` - Include timestamps (default: false)
- `log_previous` - Get logs from previous container instance (default: false)
- `namespace` - Target namespace (default: "default")

**kubectl Equivalent:**
```bash
kubectl logs <pod> -c <container> --tail=<lines> -f --timestamps --previous -n <namespace>
```

**Example:**
```yaml
- name: Get application logs
  module: cloud/kubernetes
  vars:
    command: get_pod_logs
    pod_name: webapp-abc123
    container_name: app
    log_tail_lines: 200
    log_timestamps: true
    namespace: production
```

---

### exec_pod

Execute a command in a running pod.

**Required Parameters:**
- `pod_name` - Name of the pod
- `exec_command` - Command to execute (as array)

**Optional Parameters:**
- `container_name` - Specific container name (for multi-container pods)
- `exec_stdin` - Enable stdin (default: false)
- `exec_tty` - Allocate TTY (default: false)
- `namespace` - Target namespace (default: "default")

**kubectl Equivalent:**
```bash
kubectl exec <pod> -c <container> -i -t -- <command>
```

**Example:**
```yaml
- name: Run migration
  module: cloud/kubernetes
  vars:
    command: exec_pod
    pod_name: webapp-abc123
    container_name: app
    namespace: production
    exec_command:
      - /bin/sh
      - -c
      - python manage.py migrate
```

---

## ConfigMap Management

### create_configmap

Create a ConfigMap from literal values or files.

**Required Parameters:**
- `configmap_name` - Name of the ConfigMap

**Optional Parameters:**
- `configmap_data` - Key-value pairs (object)
- `configmap_from_file` - Path to file
- `namespace` - Target namespace (default: "default")

**kubectl Equivalent:**
```bash
kubectl create configmap <name> --from-literal=key=value --from-file=<path> -n <namespace>
```

**Example:**
```yaml
- name: Create app config
  module: cloud/kubernetes
  vars:
    command: create_configmap
    configmap_name: app-config
    namespace: production
    configmap_data:
      database_url: postgres://db:5432/app
      cache_ttl: "3600"
      log_level: info
```

---

### delete_configmap

Delete a ConfigMap.

**Required Parameters:**
- `configmap_name` - Name of the ConfigMap

**Optional Parameters:**
- `namespace` - Target namespace (default: "default")

**kubectl Equivalent:**
```bash
kubectl delete configmap <name> -n <namespace>
```

**Example:**
```yaml
- name: Remove old config
  module: cloud/kubernetes
  vars:
    command: delete_configmap
    configmap_name: old-config
    namespace: production
```

---

### list_configmaps

List all ConfigMaps in a namespace.

**Optional Parameters:**
- `namespace` - Target namespace (default: "default")

**kubectl Equivalent:**
```bash
kubectl get configmaps -n <namespace>
```

**Example:**
```yaml
- name: List all ConfigMaps
  module: cloud/kubernetes
  vars:
    command: list_configmaps
    namespace: production
```

---

## Secret Management

### create_secret

Create a Secret from literal values or files.

**Required Parameters:**
- `secret_name` - Name of the Secret

**Optional Parameters:**
- `secret_type` - Type of secret (default: "generic")
  - generic (Opaque)
  - kubernetes.io/tls
  - kubernetes.io/dockerconfigjson
- `secret_data` - Key-value pairs (object)
- `secret_from_file` - Path to file
- `namespace` - Target namespace (default: "default")

**kubectl Equivalent:**
```bash
kubectl create secret <type> <name> --from-literal=key=value --from-file=<path> -n <namespace>
```

**Example:**
```yaml
- name: Create database credentials
  module: cloud/kubernetes
  vars:
    command: create_secret
    secret_name: db-creds
    secret_type: generic
    namespace: production
    secret_data:
      username: admin
      password: super_secret_password
```

---

### delete_secret

Delete a Secret.

**Required Parameters:**
- `secret_name` - Name of the Secret

**Optional Parameters:**
- `namespace` - Target namespace (default: "default")

**kubectl Equivalent:**
```bash
kubectl delete secret <name> -n <namespace>
```

**Example:**
```yaml
- name: Remove old secret
  module: cloud/kubernetes
  vars:
    command: delete_secret
    secret_name: old-creds
    namespace: production
```

---

### list_secrets

List all Secrets in a namespace.

**Optional Parameters:**
- `namespace` - Target namespace (default: "default")

**kubectl Equivalent:**
```bash
kubectl get secrets -n <namespace>
```

**Example:**
```yaml
- name: List all Secrets
  module: cloud/kubernetes
  vars:
    command: list_secrets
    namespace: production
```

---

## Namespace Management

### create_namespace

Create a new namespace.

**Required Parameters:**
- `namespace_name` - Name of the namespace

**kubectl Equivalent:**
```bash
kubectl create namespace <name>
```

**Example:**
```yaml
- name: Create production namespace
  module: cloud/kubernetes
  vars:
    command: create_namespace
    namespace_name: production
```

---

### delete_namespace

Delete a namespace (deletes all resources in it).

**Required Parameters:**
- `namespace_name` - Name of the namespace

**kubectl Equivalent:**
```bash
kubectl delete namespace <name>
```

**Example:**
```yaml
- name: Remove old namespace
  module: cloud/kubernetes
  vars:
    command: delete_namespace
    namespace_name: old-environment
```

---

### list_namespaces

List all namespaces in the cluster.

**kubectl Equivalent:**
```bash
kubectl get namespaces
```

**Example:**
```yaml
- name: List all namespaces
  module: cloud/kubernetes
  vars:
    command: list_namespaces
```

---

## Ingress Management

### create_ingress

Create an Ingress resource for HTTP/HTTPS routing.

**Required Parameters:**
- `ingress_name` - Name of the Ingress
- `ingress_host` - Hostname for routing
- `ingress_service_name` - Backend service name

**Optional Parameters:**
- `ingress_path` - Path for routing (default: "/")
- `ingress_service_port` - Backend service port (default: 80)
- `namespace` - Target namespace (default: "default")

**kubectl Equivalent:**
```bash
kubectl create ingress <name> --rule="<host><path>=<service>:<port>" -n <namespace>
```

**Example:**
```yaml
- name: Create ingress
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

---

### delete_ingress

Delete an Ingress resource.

**Required Parameters:**
- `ingress_name` - Name of the Ingress

**Optional Parameters:**
- `namespace` - Target namespace (default: "default")

**kubectl Equivalent:**
```bash
kubectl delete ingress <name> -n <namespace>
```

**Example:**
```yaml
- name: Remove ingress
  module: cloud/kubernetes
  vars:
    command: delete_ingress
    ingress_name: webapp-ingress
    namespace: production
```

---

### list_ingress

List all Ingress resources in a namespace.

**Optional Parameters:**
- `namespace` - Target namespace (default: "default")

**kubectl Equivalent:**
```bash
kubectl get ingress -n <namespace>
```

**Example:**
```yaml
- name: List all ingress resources
  module: cloud/kubernetes
  vars:
    command: list_ingress
    namespace: production
```

---

### describe_ingress

Get detailed information about an Ingress resource.

**Required Parameters:**
- `ingress_name` - Name of the Ingress

**Optional Parameters:**
- `namespace` - Target namespace (default: "default")

**kubectl Equivalent:**
```bash
kubectl describe ingress <name> -n <namespace>
```

**Example:**
```yaml
- name: Get ingress details
  module: cloud/kubernetes
  vars:
    command: describe_ingress
    ingress_name: webapp-ingress
    namespace: production
```

---

## PersistentVolumeClaim Management

### create_pvc

Create a PersistentVolumeClaim for storage.

**Required Parameters:**
- `pvc_name` - Name of the PVC

**Optional Parameters:**
- `storage_size` - Size of storage (default: "10Gi")
- `storage_class` - StorageClass name
- `access_mode` - Access mode (default: "ReadWriteOnce")
  - ReadWriteOnce
  - ReadOnlyMany
  - ReadWriteMany
- `namespace` - Target namespace (default: "default")

**kubectl Equivalent:**
```bash
# Created via manifest
```

**Example:**
```yaml
- name: Create database storage
  module: cloud/kubernetes
  vars:
    command: create_pvc
    pvc_name: db-storage
    storage_size: 50Gi
    storage_class: fast-ssd
    access_mode: ReadWriteOnce
    namespace: production
```

---

### delete_pvc

Delete a PersistentVolumeClaim.

**Required Parameters:**
- `pvc_name` - Name of the PVC

**Optional Parameters:**
- `namespace` - Target namespace (default: "default")

**kubectl Equivalent:**
```bash
kubectl delete pvc <name> -n <namespace>
```

**Example:**
```yaml
- name: Remove storage
  module: cloud/kubernetes
  vars:
    command: delete_pvc
    pvc_name: old-storage
    namespace: production
```

---

### list_pvcs

List all PersistentVolumeClaims in a namespace.

**Optional Parameters:**
- `namespace` - Target namespace (default: "default")

**kubectl Equivalent:**
```bash
kubectl get pvc -n <namespace>
```

**Example:**
```yaml
- name: List all PVCs
  module: cloud/kubernetes
  vars:
    command: list_pvcs
    namespace: production
```

---

### describe_pvc

Get detailed information about a PVC.

**Required Parameters:**
- `pvc_name` - Name of the PVC

**Optional Parameters:**
- `namespace` - Target namespace (default: "default")

**kubectl Equivalent:**
```bash
kubectl describe pvc <name> -n <namespace>
```

**Example:**
```yaml
- name: Get PVC details
  module: cloud/kubernetes
  vars:
    command: describe_pvc
    pvc_name: db-storage
    namespace: production
```

---

## Additional Operations

### apply_manifest

Apply a Kubernetes manifest from file or inline content.

**Required Parameters:**
- Either `manifest_file` OR `manifest_content`

**Optional Parameters:**
- `namespace` - Target namespace (default: "default")

**kubectl Equivalent:**
```bash
kubectl apply -f <file>
```

**Example (from file):**
```yaml
- name: Apply deployment manifest
  module: cloud/kubernetes
  vars:
    command: apply_manifest
    manifest_file: /path/to/deployment.yaml
    namespace: production
```

**Example (inline content):**
```yaml
- name: Create custom resource
  module: cloud/kubernetes
  vars:
    command: apply_manifest
    namespace: production
    manifest_content: |
      apiVersion: apps/v1
      kind: Deployment
      metadata:
        name: webapp
      spec:
        replicas: 3
        selector:
          matchLabels:
            app: webapp
        template:
          metadata:
            labels:
              app: webapp
          spec:
            containers:
            - name: webapp
              image: nginx:latest
              ports:
              - containerPort: 80
```

---

### delete_manifest

Delete resources defined in a manifest file.

**Required Parameters:**
- `manifest_file` - Path to manifest file

**kubectl Equivalent:**
```bash
kubectl delete -f <file>
```

**Example:**
```yaml
- name: Delete resources from manifest
  module: cloud/kubernetes
  vars:
    command: delete_manifest
    manifest_file: /path/to/deployment.yaml
```

---

### get_resource

Get any Kubernetes resource type.

**Required Parameters:**
- `resource_type` - Type of resource (pods, services, deployments, etc.)

**Optional Parameters:**
- `resource_name` - Specific resource name
- `namespace` - Target namespace (default: "default")
- `output_format` - Output format (default: "wide")
  - wide
  - json
  - yaml
  - name

**kubectl Equivalent:**
```bash
kubectl get <type> <name> -o <format> -n <namespace>
```

**Example:**
```yaml
- name: Get all deployments
  module: cloud/kubernetes
  vars:
    command: get_resource
    resource_type: deployments
    namespace: production
    output_format: wide
```

**Example (specific resource as JSON):**
```yaml
- name: Get deployment as JSON
  module: cloud/kubernetes
  vars:
    command: get_resource
    resource_type: deployment
    resource_name: webapp
    namespace: production
    output_format: json
```

---

## Common Parameters

All commands support these common parameters:

- `namespace` - Target namespace (default: "default")
- `kubeconfig` - Path to kubeconfig file (optional)
- `context` - Kubernetes context to use (optional)
- `dry_run` - Perform dry-run (default: false)
- `wait` - Wait for operation to complete (default: true)
- `timeout` - Operation timeout (default: "5m")
- `output_format` - Output format for get/list commands

## Exit Status

The module returns one of these status values:

- `ok` - Operation completed successfully, no changes made
- `changed` - Operation completed successfully, changes were made
- `failed` - Operation failed

## Output Facts

Common facts returned by operations:

- `output` - Raw kubectl command output
- `command` - Full kubectl command executed
- `error` - Error message (if failed)
- `resource_created` - Name of created resource
- `resource_name` - Name of affected resource

## Notes

1. All commands require `kubectl` to be installed on the target host
2. kubeconfig must be properly configured or in-cluster credentials available
3. RBAC permissions must allow the requested operations
4. Resource names must follow Kubernetes naming conventions (lowercase, alphanumeric, hyphens)
5. Namespace must exist before creating resources in it
6. Some operations may take time to complete (use `wait: true`)
