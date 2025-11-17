# OpenFroyo Cloud Platform Modules

This directory contains modules for managing cloud platforms and container orchestration.

## Available Modules (2 Total)

### 1. aws (Amazon Web Services)
**Purpose:** Complete AWS cloud infrastructure management

**Key Features:**
- 40 comprehensive operations
- EC2 instance lifecycle management
- S3 bucket and object operations
- VPC and networking (subnets, security groups)
- IAM users, roles, and policies
- RDS database management
- Lambda serverless functions

**Compatibility:** Any AWS region with AWS CLI installed

**Size:** 495KB

**Authentication:** AWS profiles, environment variables, or IAM roles

---

### 2. kubernetes (Kubernetes Cluster Management)
**Purpose:** Complete Kubernetes resource management

**Key Features:**
- 38+ operations
- Deployment lifecycle management
- Service discovery and load balancing
- ConfigMap and Secret management
- Pod operations (logs, exec, describe)
- Ingress and routing
- PersistentVolume claims
- Namespace management

**Compatibility:** Any Kubernetes cluster (1.19+)

**Size:** 3.6MB

**Authentication:** kubeconfig file or in-cluster config

---

## Comparison Matrix

| Feature | AWS | Kubernetes |
|---------|-----|------------|
| Compute | ✅ EC2 (10 ops) | ✅ Pods/Deployments (13 ops) |
| Storage | ✅ S3 (8 ops) | ✅ PVC (4 ops) |
| Networking | ✅ VPC/SG (8 ops) | ✅ Service/Ingress (9 ops) |
| Config Management | ❌ | ✅ ConfigMap/Secret (6 ops) |
| IAM/RBAC | ✅ (6 ops) | ✅ (via kubectl) |
| Databases | ✅ RDS (5 ops) | ❌ |
| Serverless | ✅ Lambda (3 ops) | ❌ |
| Namespaces | ❌ | ✅ (3 ops) |

## When to Use Each Module

### Use **aws** when:
- Managing AWS cloud infrastructure
- Provisioning EC2 instances
- Managing S3 storage buckets
- Creating VPCs and networking
- Setting up RDS databases
- Deploying Lambda functions
- Managing IAM users and roles

### Use **kubernetes** when:
- Deploying containerized applications
- Managing microservices architectures
- Setting up service discovery
- Configuring ingress/load balancing
- Managing application configurations
- Orchestrating container workloads
- Running cloud-native applications

## Common Usage Patterns

### AWS Infrastructure Provisioning

```yaml
# Create VPC and subnet
- module: cloud/aws
  vars:
    aws_region: us-east-1
    command: create_vpc
    cidr_block: 10.0.0.0/16

- module: cloud/aws
  vars:
    command: create_subnet
    vpc_id: "{{ vpc_id }}"
    cidr_block: 10.0.1.0/24

# Launch EC2 instance
- module: cloud/aws
  vars:
    command: create_instance
    ami_id: ami-0c55b159cbfafe1f0
    instance_type: t3.medium
    subnet_id: "{{ subnet_id }}"
    key_name: my-keypair
```

### Kubernetes Application Deployment

```yaml
# Create deployment
- module: cloud/kubernetes
  vars:
    namespace: production
    command: create_deployment
    deployment_name: web-app
    image: nginx:latest
    replicas: 3
    port: 80

# Expose via service
- module: cloud/kubernetes
  vars:
    command: create_service
    service_name: web-app-service
    service_type: LoadBalancer
    port: 80
    target_port: 80

# Create ingress
- module: cloud/kubernetes
  vars:
    command: create_ingress
    ingress_name: web-app-ingress
    host: app.example.com
    path: /
    service_name: web-app-service
    service_port: 80
```

## Multi-Cloud Scenarios

### AWS + Kubernetes Hybrid

```yaml
# Provision AWS RDS database
- module: cloud/aws
  vars:
    command: create_db_instance
    db_instance_identifier: app-db
    engine: postgres
    db_instance_class: db.t3.medium

# Deploy app on Kubernetes with RDS connection
- module: cloud/kubernetes
  vars:
    command: create_secret
    secret_name: db-credentials
    data:
      DB_HOST: "{{ rds_endpoint }}"
      DB_PASSWORD: "{{ db_password }}"

- module: cloud/kubernetes
  vars:
    command: create_deployment
    deployment_name: backend-app
    image: myapp:latest
    env_from_secret: db-credentials
```

## Security Best Practices

### AWS Security
- Use IAM roles instead of access keys when possible
- Enable MFA for production accounts
- Use least-privilege IAM policies
- Encrypt S3 buckets at rest
- Use security groups restrictively
- Enable VPC Flow Logs
- Rotate access keys regularly

### Kubernetes Security
- Use namespaces for isolation
- Implement RBAC for access control
- Use NetworkPolicies for pod communication
- Store sensitive data in Secrets (encrypted at rest)
- Use PodSecurityPolicies
- Enable audit logging
- Scan container images for vulnerabilities

## Performance Considerations

### AWS
- Choose appropriate instance types for workload
- Use EC2 Auto Scaling for variable load
- Enable S3 Transfer Acceleration for large files
- Use ElastiCache for database caching
- Monitor with CloudWatch
- Use Reserved Instances for cost savings

### Kubernetes
- Set resource requests and limits
- Use HorizontalPodAutoscaler for scaling
- Implement liveness and readiness probes
- Use PersistentVolumes for stateful apps
- Monitor with Prometheus/Grafana
- Optimize container images (multi-stage builds)

## Cost Optimization

### AWS
- Use spot instances for non-critical workloads
- Stop unused EC2 instances
- Use S3 lifecycle policies
- Delete unused EBS volumes
- Use Reserved Instances for predictable workloads
- Set up billing alerts
- Tag resources for cost tracking

### Kubernetes
- Set resource limits to prevent overprovisioning
- Use cluster autoscaling
- Implement pod disruption budgets
- Use node affinity for workload placement
- Monitor resource utilization
- Right-size workloads

## Troubleshooting

### AWS Issues

**Connection Problems:**
- Verify AWS CLI is installed: `aws --version`
- Check AWS credentials: `aws sts get-caller-identity`
- Verify region configuration
- Check network connectivity to AWS endpoints

**Permission Errors:**
- Review IAM policy permissions
- Check for explicit deny statements
- Verify resource-based policies
- Review service control policies (SCPs)

### Kubernetes Issues

**Connection Problems:**
- Verify kubectl is installed: `kubectl version`
- Check kubeconfig: `kubectl config view`
- Test cluster connection: `kubectl cluster-info`
- Verify context: `kubectl config current-context`

**Deployment Issues:**
- Check pod status: `kubectl get pods -n <namespace>`
- View pod logs: `kubectl logs <pod-name>`
- Describe pod: `kubectl describe pod <pod-name>`
- Check events: `kubectl get events -n <namespace>`

## Documentation

Each module includes comprehensive documentation:
- **README.md** - Complete usage guide
- **COMMANDS.md** - Detailed command reference
- **WORKFLOWS.md** - Real-world scenarios
- **test.ofy** - Example test stack

## Module Statistics

- **2 cloud modules**
- **78+ total operations**
- **~4.1MB total WASM binaries** (source only, binaries in .gitignore)
- **Multi-cloud support** (AWS public cloud + Kubernetes orchestration)
- **Production-ready** with comprehensive documentation

## Future Enhancements

Potential additions:
- Azure module (VMs, Storage, AKS)
- Google Cloud Platform (GCE, GCS, GKE)
- DigitalOcean module
- Terraform integration
- CloudFormation support
- Helm chart deployment
