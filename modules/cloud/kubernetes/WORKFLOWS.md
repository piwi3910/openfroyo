# Kubernetes Module - Workflow Examples

Real-world deployment scenarios and best practices using the Kubernetes module.

## Table of Contents

- [Basic Web Application Deployment](#basic-web-application-deployment)
- [Microservices Architecture](#microservices-architecture)
- [Stateful Application with Database](#stateful-application-with-database)
- [Blue-Green Deployment](#blue-green-deployment)
- [Canary Deployment](#canary-deployment)
- [Multi-Tier Application Stack](#multi-tier-application-stack)
- [CI/CD Pipeline Integration](#cicd-pipeline-integration)
- [Disaster Recovery Setup](#disaster-recovery-setup)
- [Security Hardening](#security-hardening)
- [Monitoring and Logging Setup](#monitoring-and-logging-setup)

---

## Basic Web Application Deployment

Deploy a simple web application with load balancer.

```yaml
---
inventory: "@k8s-cluster"

defaults:
  namespace: webapp
  kubeconfig: ~/.kube/config

run:
  # 1. Create dedicated namespace
  - name: Create application namespace
    module: cloud/kubernetes
    vars:
      command: create_namespace
      namespace_name: webapp

  # 2. Create ConfigMap for application settings
  - name: Create application configuration
    module: cloud/kubernetes
    vars:
      command: create_configmap
      configmap_name: webapp-config
      configmap_data:
        LOG_LEVEL: info
        PORT: "8080"
        ENV: production

  # 3. Create Secret for sensitive data
  - name: Create application secrets
    module: cloud/kubernetes
    vars:
      command: create_secret
      secret_name: webapp-secrets
      secret_type: generic
      secret_data:
        API_KEY: your-api-key-here
        SESSION_SECRET: random-session-secret

  # 4. Deploy the application
  - name: Deploy web application
    module: cloud/kubernetes
    vars:
      command: create_deployment
      deployment_name: webapp
      image: webapp:v1.0.0
      replicas: 3
      port: 8080

  # 5. Expose via LoadBalancer
  - name: Expose application
    module: cloud/kubernetes
    vars:
      command: expose_deployment
      deployment_name: webapp
      service_type: LoadBalancer
      port: 80

  # 6. Wait for rollout
  - name: Wait for deployment
    module: cloud/kubernetes
    vars:
      command: rollout_status
      deployment_name: webapp

  # 7. Get service details
  - name: Get service endpoint
    module: cloud/kubernetes
    vars:
      command: describe_service
      service_name: webapp
```

---

## Microservices Architecture

Deploy multiple interconnected microservices.

```yaml
---
inventory: "@k8s-cluster"

defaults:
  namespace: microservices
  image_registry: registry.example.com
  replicas: 2

run:
  # Setup namespace
  - name: Create microservices namespace
    module: cloud/kubernetes
    vars:
      command: create_namespace
      namespace_name: microservices

  # Shared configuration
  - name: Create shared config
    module: cloud/kubernetes
    vars:
      command: create_configmap
      configmap_name: shared-config
      configmap_data:
        ENVIRONMENT: production
        LOG_FORMAT: json
        TRACE_ENABLED: "true"

  # API Gateway
  - name: Deploy API Gateway
    module: cloud/kubernetes
    vars:
      command: create_deployment
      deployment_name: api-gateway
      image: "{{ image_registry }}/api-gateway:latest"
      replicas: 3
      port: 8080

  - name: Expose API Gateway
    module: cloud/kubernetes
    vars:
      command: create_service
      service_name: api-gateway
      service_type: LoadBalancer
      port: 80
      target_port: 8080

  # User Service
  - name: Deploy User Service
    module: cloud/kubernetes
    vars:
      command: create_deployment
      deployment_name: user-service
      image: "{{ image_registry }}/user-service:latest"
      replicas: "{{ replicas }}"
      port: 8081

  - name: Expose User Service
    module: cloud/kubernetes
    vars:
      command: create_service
      service_name: user-service
      service_type: ClusterIP
      port: 8081
      target_port: 8081

  # Product Service
  - name: Deploy Product Service
    module: cloud/kubernetes
    vars:
      command: create_deployment
      deployment_name: product-service
      image: "{{ image_registry }}/product-service:latest"
      replicas: "{{ replicas }}"
      port: 8082

  - name: Expose Product Service
    module: cloud/kubernetes
    vars:
      command: create_service
      service_name: product-service
      service_type: ClusterIP
      port: 8082
      target_port: 8082

  # Order Service
  - name: Deploy Order Service
    module: cloud/kubernetes
    vars:
      command: create_deployment
      deployment_name: order-service
      image: "{{ image_registry }}/order-service:latest"
      replicas: "{{ replicas }}"
      port: 8083

  - name: Expose Order Service
    module: cloud/kubernetes
    vars:
      command: create_service
      service_name: order-service
      service_type: ClusterIP
      port: 8083
      target_port: 8083

  # Ingress for routing
  - name: Create ingress for API
    module: cloud/kubernetes
    vars:
      command: create_ingress
      ingress_name: api-ingress
      ingress_host: api.example.com
      ingress_path: /
      ingress_service_name: api-gateway
      ingress_service_port: 80
```

---

## Stateful Application with Database

Deploy a database with persistent storage.

```yaml
---
inventory: "@k8s-cluster"

defaults:
  namespace: database
  db_name: postgres
  storage_size: 100Gi

run:
  # 1. Create namespace
  - name: Create database namespace
    module: cloud/kubernetes
    vars:
      command: create_namespace
      namespace_name: database

  # 2. Create storage
  - name: Create persistent storage
    module: cloud/kubernetes
    vars:
      command: create_pvc
      pvc_name: "{{ db_name }}-storage"
      storage_size: "{{ storage_size }}"
      storage_class: fast-ssd
      access_mode: ReadWriteOnce

  # 3. Create database credentials
  - name: Create database secrets
    module: cloud/kubernetes
    vars:
      command: create_secret
      secret_name: "{{ db_name }}-secrets"
      secret_type: generic
      secret_data:
        POSTGRES_USER: dbadmin
        POSTGRES_PASSWORD: super_secure_password
        POSTGRES_DB: production

  # 4. Deploy database using manifest
  - name: Deploy PostgreSQL StatefulSet
    module: cloud/kubernetes
    vars:
      command: apply_manifest
      manifest_content: |
        apiVersion: apps/v1
        kind: StatefulSet
        metadata:
          name: {{ db_name }}
          namespace: {{ namespace }}
        spec:
          serviceName: {{ db_name }}
          replicas: 1
          selector:
            matchLabels:
              app: {{ db_name }}
          template:
            metadata:
              labels:
                app: {{ db_name }}
            spec:
              containers:
              - name: postgres
                image: postgres:14
                ports:
                - containerPort: 5432
                  name: postgres
                envFrom:
                - secretRef:
                    name: {{ db_name }}-secrets
                volumeMounts:
                - name: data
                  mountPath: /var/lib/postgresql/data
              volumes:
              - name: data
                persistentVolumeClaim:
                  claimName: {{ db_name }}-storage

  # 5. Create headless service for StatefulSet
  - name: Create database service
    module: cloud/kubernetes
    vars:
      command: apply_manifest
      manifest_content: |
        apiVersion: v1
        kind: Service
        metadata:
          name: {{ db_name }}
          namespace: {{ namespace }}
        spec:
          clusterIP: None
          selector:
            app: {{ db_name }}
          ports:
          - port: 5432
            targetPort: 5432

  # 6. Verify deployment
  - name: Check database pod
    module: cloud/kubernetes
    vars:
      command: list_pods

  - name: Get database logs
    module: cloud/kubernetes
    vars:
      command: get_pod_logs
      pod_name: "{{ db_name }}-0"
      log_tail_lines: 50
```

---

## Blue-Green Deployment

Zero-downtime deployment using blue-green strategy.

```yaml
---
inventory: "@k8s-cluster"

defaults:
  namespace: production
  app_name: webapp
  blue_version: v1.0.0
  green_version: v2.0.0

run:
  # Assume blue (v1.0.0) is already running

  # 1. Deploy green version
  - name: Deploy green version
    module: cloud/kubernetes
    vars:
      command: create_deployment
      deployment_name: "{{ app_name }}-green"
      image: "webapp:{{ green_version }}"
      replicas: 3
      port: 8080

  # 2. Wait for green to be ready
  - name: Wait for green deployment
    module: cloud/kubernetes
    vars:
      command: rollout_status
      deployment_name: "{{ app_name }}-green"

  # 3. Test green deployment (internal service)
  - name: Create test service for green
    module: cloud/kubernetes
    vars:
      command: apply_manifest
      manifest_content: |
        apiVersion: v1
        kind: Service
        metadata:
          name: {{ app_name }}-green-test
          namespace: {{ namespace }}
        spec:
          type: ClusterIP
          selector:
            app: {{ app_name }}-green
          ports:
          - port: 8080
            targetPort: 8080

  # Manual testing phase here...

  # 4. Switch traffic to green (update main service selector)
  - name: Switch traffic to green
    module: cloud/kubernetes
    vars:
      command: apply_manifest
      manifest_content: |
        apiVersion: v1
        kind: Service
        metadata:
          name: {{ app_name }}
          namespace: {{ namespace }}
        spec:
          type: LoadBalancer
          selector:
            app: {{ app_name }}-green
          ports:
          - port: 80
            targetPort: 8080

  # 5. Monitor for issues
  - name: Check green deployment status
    module: cloud/kubernetes
    vars:
      command: describe_deployment
      deployment_name: "{{ app_name }}-green"

  # 6. Scale down blue (keep for rollback)
  - name: Scale down blue deployment
    module: cloud/kubernetes
    vars:
      command: scale_deployment
      deployment_name: "{{ app_name }}-blue"
      replicas: 1

  # After verification period, delete blue:
  # - name: Delete blue deployment
  #   module: cloud/kubernetes
  #   vars:
  #     command: delete_deployment
  #     deployment_name: "{{ app_name }}-blue"
```

---

## Canary Deployment

Gradual rollout with traffic splitting.

```yaml
---
inventory: "@k8s-cluster"

defaults:
  namespace: production
  app_name: webapp
  stable_version: v1.0.0
  canary_version: v2.0.0

run:
  # Assume stable version is running with 10 replicas

  # 1. Deploy canary with 1 replica (10% traffic)
  - name: Deploy canary version
    module: cloud/kubernetes
    vars:
      command: create_deployment
      deployment_name: "{{ app_name }}-canary"
      image: "webapp:{{ canary_version }}"
      replicas: 1
      port: 8080

  # 2. Wait for canary
  - name: Wait for canary deployment
    module: cloud/kubernetes
    vars:
      command: rollout_status
      deployment_name: "{{ app_name }}-canary"

  # 3. Update service to include both deployments
  - name: Configure service for canary
    module: cloud/kubernetes
    vars:
      command: apply_manifest
      manifest_content: |
        apiVersion: v1
        kind: Service
        metadata:
          name: {{ app_name }}
          namespace: {{ namespace }}
        spec:
          type: LoadBalancer
          selector:
            app: {{ app_name }}  # Both deployments have this label
          ports:
          - port: 80
            targetPort: 8080

  # Monitor metrics, errors, performance...

  # 4. Gradually increase canary (20% traffic)
  - name: Scale canary to 20%
    module: cloud/kubernetes
    vars:
      command: scale_deployment
      deployment_name: "{{ app_name }}-canary"
      replicas: 2

  # 5. Continue monitoring...

  # 6. Scale canary to 50%
  - name: Scale canary to 50%
    module: cloud/kubernetes
    vars:
      command: scale_deployment
      deployment_name: "{{ app_name }}-canary"
      replicas: 5

  - name: Scale stable to 50%
    module: cloud/kubernetes
    vars:
      command: scale_deployment
      deployment_name: "{{ app_name }}-stable"
      replicas: 5

  # 7. Full rollout - update stable deployment
  - name: Update stable deployment
    module: cloud/kubernetes
    vars:
      command: update_deployment
      deployment_name: "{{ app_name }}-stable"
      image: "webapp:{{ canary_version }}"

  - name: Scale stable to full
    module: cloud/kubernetes
    vars:
      command: scale_deployment
      deployment_name: "{{ app_name }}-stable"
      replicas: 10

  # 8. Remove canary
  - name: Delete canary deployment
    module: cloud/kubernetes
    vars:
      command: delete_deployment
      deployment_name: "{{ app_name }}-canary"
```

---

## Multi-Tier Application Stack

Full stack deployment (frontend, backend, database, cache).

```yaml
---
inventory: "@k8s-cluster"

defaults:
  namespace: app-stack
  environment: production

run:
  # 1. Setup namespace
  - name: Create namespace
    module: cloud/kubernetes
    vars:
      command: create_namespace
      namespace_name: app-stack

  # 2. Deploy Redis cache
  - name: Deploy Redis
    module: cloud/kubernetes
    vars:
      command: create_deployment
      deployment_name: redis
      image: redis:7-alpine
      replicas: 1
      port: 6379

  - name: Expose Redis
    module: cloud/kubernetes
    vars:
      command: create_service
      service_name: redis
      service_type: ClusterIP
      port: 6379

  # 3. Deploy PostgreSQL database
  - name: Create database storage
    module: cloud/kubernetes
    vars:
      command: create_pvc
      pvc_name: postgres-storage
      storage_size: 50Gi
      access_mode: ReadWriteOnce

  - name: Create database secrets
    module: cloud/kubernetes
    vars:
      command: create_secret
      secret_name: postgres-secrets
      secret_data:
        POSTGRES_USER: appuser
        POSTGRES_PASSWORD: securepassword
        POSTGRES_DB: appdb

  - name: Deploy PostgreSQL
    module: cloud/kubernetes
    vars:
      command: create_deployment
      deployment_name: postgres
      image: postgres:14
      replicas: 1
      port: 5432

  - name: Expose PostgreSQL
    module: cloud/kubernetes
    vars:
      command: create_service
      service_name: postgres
      service_type: ClusterIP
      port: 5432

  # 4. Deploy Backend API
  - name: Create backend config
    module: cloud/kubernetes
    vars:
      command: create_configmap
      configmap_name: backend-config
      configmap_data:
        DATABASE_HOST: postgres
        REDIS_HOST: redis
        PORT: "8080"

  - name: Deploy backend
    module: cloud/kubernetes
    vars:
      command: create_deployment
      deployment_name: backend
      image: backend-api:latest
      replicas: 3
      port: 8080

  - name: Expose backend
    module: cloud/kubernetes
    vars:
      command: create_service
      service_name: backend
      service_type: ClusterIP
      port: 8080

  # 5. Deploy Frontend
  - name: Create frontend config
    module: cloud/kubernetes
    vars:
      command: create_configmap
      configmap_name: frontend-config
      configmap_data:
        API_URL: http://backend:8080
        ENVIRONMENT: "{{ environment }}"

  - name: Deploy frontend
    module: cloud/kubernetes
    vars:
      command: create_deployment
      deployment_name: frontend
      image: frontend-app:latest
      replicas: 2
      port: 80

  - name: Expose frontend
    module: cloud/kubernetes
    vars:
      command: create_service
      service_name: frontend
      service_type: LoadBalancer
      port: 80

  # 6. Setup Ingress
  - name: Create ingress rules
    module: cloud/kubernetes
    vars:
      command: apply_manifest
      manifest_content: |
        apiVersion: networking.k8s.io/v1
        kind: Ingress
        metadata:
          name: app-ingress
          namespace: {{ namespace }}
          annotations:
            nginx.ingress.kubernetes.io/rewrite-target: /
        spec:
          rules:
          - host: app.example.com
            http:
              paths:
              - path: /
                pathType: Prefix
                backend:
                  service:
                    name: frontend
                    port:
                      number: 80
              - path: /api
                pathType: Prefix
                backend:
                  service:
                    name: backend
                    port:
                      number: 8080

  # 7. Verify deployment
  - name: List all resources
    module: cloud/kubernetes
    vars:
      command: list_pods
```

---

## CI/CD Pipeline Integration

Automated deployment from CI/CD pipeline.

```yaml
---
# This stack is triggered by CI/CD after successful build

inventory: "@k8s-cluster"

defaults:
  namespace: "{{ ci_namespace | default('staging') }}"
  image_tag: "{{ ci_commit_sha }}"
  app_name: webapp

run:
  # 1. Update deployment with new image
  - name: Deploy new version
    module: cloud/kubernetes
    vars:
      command: update_deployment
      deployment_name: "{{ app_name }}"
      image: "registry.example.com/{{ app_name }}:{{ image_tag }}"

  # 2. Wait for rollout
  - name: Wait for rollout to complete
    module: cloud/kubernetes
    vars:
      command: rollout_status
      deployment_name: "{{ app_name }}"
      timeout: 10m

  # 3. Run smoke tests (via job)
  - name: Run smoke tests
    module: cloud/kubernetes
    vars:
      command: apply_manifest
      manifest_content: |
        apiVersion: batch/v1
        kind: Job
        metadata:
          name: smoke-tests-{{ ci_commit_sha[:8] }}
          namespace: {{ namespace }}
        spec:
          template:
            spec:
              containers:
              - name: test-runner
                image: smoke-tests:latest
                env:
                - name: APP_URL
                  value: http://{{ app_name }}:8080
              restartPolicy: Never
          backoffLimit: 2

  # 4. Check test results
  - name: Get test logs
    module: cloud/kubernetes
    vars:
      command: get_pod_logs
      pod_name: "smoke-tests-{{ ci_commit_sha[:8] }}"
      log_follow: true

  # 5. If tests pass, tag as production-ready
  - name: Label deployment as tested
    module: cloud/kubernetes
    vars:
      command: apply_manifest
      manifest_content: |
        apiVersion: apps/v1
        kind: Deployment
        metadata:
          name: {{ app_name }}
          namespace: {{ namespace }}
          labels:
            tested: "true"
            commit: "{{ ci_commit_sha }}"
            timestamp: "{{ ci_timestamp }}"
        spec:
          # ... existing spec
```

---

## Disaster Recovery Setup

Setup for disaster recovery and backup.

```yaml
---
inventory: "@k8s-cluster"

defaults:
  namespace: production
  backup_namespace: backups

run:
  # 1. Create backup namespace
  - name: Create backup namespace
    module: cloud/kubernetes
    vars:
      command: create_namespace
      namespace_name: "{{ backup_namespace }}"

  # 2. Setup backup storage
  - name: Create backup PVC
    module: cloud/kubernetes
    vars:
      command: create_pvc
      pvc_name: backup-storage
      namespace: "{{ backup_namespace }}"
      storage_size: 500Gi
      storage_class: standard
      access_mode: ReadWriteMany

  # 3. Create backup credentials
  - name: Create S3 backup credentials
    module: cloud/kubernetes
    vars:
      command: create_secret
      secret_name: backup-credentials
      namespace: "{{ backup_namespace }}"
      secret_data:
        AWS_ACCESS_KEY_ID: your-access-key
        AWS_SECRET_ACCESS_KEY: your-secret-key
        S3_BUCKET: disaster-recovery-backups

  # 4. Deploy backup CronJob for database
  - name: Create database backup CronJob
    module: cloud/kubernetes
    vars:
      command: apply_manifest
      manifest_content: |
        apiVersion: batch/v1
        kind: CronJob
        metadata:
          name: postgres-backup
          namespace: {{ backup_namespace }}
        spec:
          schedule: "0 2 * * *"  # Daily at 2 AM
          jobTemplate:
            spec:
              template:
                spec:
                  containers:
                  - name: backup
                    image: postgres:14
                    command:
                    - /bin/bash
                    - -c
                    - |
                      pg_dump -h postgres.production.svc.cluster.local \
                        -U $POSTGRES_USER -d $POSTGRES_DB | \
                        gzip > /backup/backup-$(date +%Y%m%d-%H%M%S).sql.gz
                    envFrom:
                    - secretRef:
                        name: postgres-secrets
                    volumeMounts:
                    - name: backup
                      mountPath: /backup
                  volumes:
                  - name: backup
                    persistentVolumeClaim:
                      claimName: backup-storage
                  restartPolicy: OnFailure

  # 5. Deploy backup sync to S3
  - name: Create S3 sync CronJob
    module: cloud/kubernetes
    vars:
      command: apply_manifest
      manifest_content: |
        apiVersion: batch/v1
        kind: CronJob
        metadata:
          name: backup-sync-s3
          namespace: {{ backup_namespace }}
        spec:
          schedule: "0 4 * * *"  # Daily at 4 AM
          jobTemplate:
            spec:
              template:
                spec:
                  containers:
                  - name: sync
                    image: amazon/aws-cli
                    command:
                    - /bin/bash
                    - -c
                    - |
                      aws s3 sync /backup s3://$S3_BUCKET/postgres-backups/ \
                        --storage-class GLACIER
                    envFrom:
                    - secretRef:
                        name: backup-credentials
                    volumeMounts:
                    - name: backup
                      mountPath: /backup
                  volumes:
                  - name: backup
                    persistentVolumeClaim:
                      claimName: backup-storage
                  restartPolicy: OnFailure

  # 6. Create restore job (manual trigger)
  - name: Create restore job template
    module: cloud/kubernetes
    vars:
      command: apply_manifest
      manifest_content: |
        apiVersion: batch/v1
        kind: Job
        metadata:
          name: postgres-restore
          namespace: {{ backup_namespace }}
        spec:
          template:
            spec:
              containers:
              - name: restore
                image: postgres:14
                command:
                - /bin/bash
                - -c
                - |
                  # Download latest backup from S3
                  aws s3 cp s3://$S3_BUCKET/postgres-backups/latest.sql.gz /tmp/
                  # Restore
                  gunzip < /tmp/latest.sql.gz | \
                    psql -h postgres.production.svc.cluster.local \
                    -U $POSTGRES_USER -d $POSTGRES_DB
                envFrom:
                - secretRef:
                    name: postgres-secrets
                - secretRef:
                    name: backup-credentials
              restartPolicy: Never
          backoffLimit: 1
```

---

## Security Hardening

Implementing security best practices.

```yaml
---
inventory: "@k8s-cluster"

defaults:
  namespace: secure-app

run:
  # 1. Create namespace with labels
  - name: Create secure namespace
    module: cloud/kubernetes
    vars:
      command: apply_manifest
      manifest_content: |
        apiVersion: v1
        kind: Namespace
        metadata:
          name: {{ namespace }}
          labels:
            security-level: high
            pod-security.kubernetes.io/enforce: restricted

  # 2. Create NetworkPolicy (deny all by default)
  - name: Create default deny network policy
    module: cloud/kubernetes
    vars:
      command: apply_manifest
      manifest_content: |
        apiVersion: networking.k8s.io/v1
        kind: NetworkPolicy
        metadata:
          name: default-deny-all
          namespace: {{ namespace }}
        spec:
          podSelector: {}
          policyTypes:
          - Ingress
          - Egress

  # 3. Create NetworkPolicy for app
  - name: Create app network policy
    module: cloud/kubernetes
    vars:
      command: apply_manifest
      manifest_content: |
        apiVersion: networking.k8s.io/v1
        kind: NetworkPolicy
        metadata:
          name: webapp-netpol
          namespace: {{ namespace }}
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
            ports:
            - protocol: TCP
              port: 8080
          egress:
          - to:
            - podSelector:
                matchLabels:
                  app: database
            ports:
            - protocol: TCP
              port: 5432
          - to:  # Allow DNS
            - namespaceSelector:
                matchLabels:
                  name: kube-system
            ports:
            - protocol: UDP
              port: 53

  # 4. Create service account
  - name: Create app service account
    module: cloud/kubernetes
    vars:
      command: apply_manifest
      manifest_content: |
        apiVersion: v1
        kind: ServiceAccount
        metadata:
          name: webapp-sa
          namespace: {{ namespace }}

  # 5. Create RBAC role
  - name: Create app role
    module: cloud/kubernetes
    vars:
      command: apply_manifest
      manifest_content: |
        apiVersion: rbac.authorization.k8s.io/v1
        kind: Role
        metadata:
          name: webapp-role
          namespace: {{ namespace }}
        rules:
        - apiGroups: [""]
          resources: ["configmaps"]
          verbs: ["get", "list"]
        - apiGroups: [""]
          resources: ["secrets"]
          verbs: ["get"]

  # 6. Create role binding
  - name: Create role binding
    module: cloud/kubernetes
    vars:
      command: apply_manifest
      manifest_content: |
        apiVersion: rbac.authorization.k8s.io/v1
        kind: RoleBinding
        metadata:
          name: webapp-binding
          namespace: {{ namespace }}
        roleRef:
          apiGroup: rbac.authorization.k8s.io
          kind: Role
          name: webapp-role
        subjects:
        - kind: ServiceAccount
          name: webapp-sa
          namespace: {{ namespace }}

  # 7. Deploy secure application
  - name: Deploy secure application
    module: cloud/kubernetes
    vars:
      command: apply_manifest
      manifest_content: |
        apiVersion: apps/v1
        kind: Deployment
        metadata:
          name: webapp
          namespace: {{ namespace }}
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
              serviceAccountName: webapp-sa
              securityContext:
                runAsNonRoot: true
                runAsUser: 1000
                fsGroup: 2000
                seccompProfile:
                  type: RuntimeDefault
              containers:
              - name: webapp
                image: webapp:secure
                securityContext:
                  allowPrivilegeEscalation: false
                  readOnlyRootFilesystem: true
                  capabilities:
                    drop:
                    - ALL
                ports:
                - containerPort: 8080
                resources:
                  limits:
                    cpu: "1000m"
                    memory: "512Mi"
                  requests:
                    cpu: "100m"
                    memory: "128Mi"
                livenessProbe:
                  httpGet:
                    path: /health
                    port: 8080
                  initialDelaySeconds: 30
                readinessProbe:
                  httpGet:
                    path: /ready
                    port: 8080
                  initialDelaySeconds: 5

  # 8. Create ResourceQuota
  - name: Create resource quota
    module: cloud/kubernetes
    vars:
      command: apply_manifest
      manifest_content: |
        apiVersion: v1
        kind: ResourceQuota
        metadata:
          name: namespace-quota
          namespace: {{ namespace }}
        spec:
          hard:
            requests.cpu: "10"
            requests.memory: 20Gi
            limits.cpu: "20"
            limits.memory: 40Gi
            persistentvolumeclaims: "5"

  # 9. Create LimitRange
  - name: Create limit range
    module: cloud/kubernetes
    vars:
      command: apply_manifest
      manifest_content: |
        apiVersion: v1
        kind: LimitRange
        metadata:
          name: resource-limits
          namespace: {{ namespace }}
        spec:
          limits:
          - max:
              cpu: "2"
              memory: 2Gi
            min:
              cpu: 100m
              memory: 128Mi
            default:
              cpu: 500m
              memory: 512Mi
            defaultRequest:
              cpu: 200m
              memory: 256Mi
            type: Container
```

---

## Monitoring and Logging Setup

Deploy monitoring and logging infrastructure.

```yaml
---
inventory: "@k8s-cluster"

defaults:
  namespace: monitoring

run:
  # 1. Create monitoring namespace
  - name: Create monitoring namespace
    module: cloud/kubernetes
    vars:
      command: create_namespace
      namespace_name: monitoring

  # 2. Deploy Prometheus
  - name: Create Prometheus config
    module: cloud/kubernetes
    vars:
      command: create_configmap
      configmap_name: prometheus-config
      configmap_data:
        prometheus.yml: |
          global:
            scrape_interval: 15s
          scrape_configs:
          - job_name: 'kubernetes-pods'
            kubernetes_sd_configs:
            - role: pod

  - name: Create Prometheus storage
    module: cloud/kubernetes
    vars:
      command: create_pvc
      pvc_name: prometheus-storage
      storage_size: 100Gi
      access_mode: ReadWriteOnce

  - name: Deploy Prometheus
    module: cloud/kubernetes
    vars:
      command: create_deployment
      deployment_name: prometheus
      image: prom/prometheus:latest
      replicas: 1
      port: 9090

  - name: Expose Prometheus
    module: cloud/kubernetes
    vars:
      command: create_service
      service_name: prometheus
      service_type: ClusterIP
      port: 9090

  # 3. Deploy Grafana
  - name: Create Grafana storage
    module: cloud/kubernetes
    vars:
      command: create_pvc
      pvc_name: grafana-storage
      storage_size: 10Gi
      access_mode: ReadWriteOnce

  - name: Create Grafana secrets
    module: cloud/kubernetes
    vars:
      command: create_secret
      secret_name: grafana-secrets
      secret_data:
        admin-user: admin
        admin-password: changeme

  - name: Deploy Grafana
    module: cloud/kubernetes
    vars:
      command: create_deployment
      deployment_name: grafana
      image: grafana/grafana:latest
      replicas: 1
      port: 3000

  - name: Expose Grafana
    module: cloud/kubernetes
    vars:
      command: create_service
      service_name: grafana
      service_type: LoadBalancer
      port: 80
      target_port: 3000

  # 4. Deploy Elasticsearch for logging
  - name: Create Elasticsearch storage
    module: cloud/kubernetes
    vars:
      command: create_pvc
      pvc_name: elasticsearch-storage
      storage_size: 200Gi
      access_mode: ReadWriteOnce

  - name: Deploy Elasticsearch
    module: cloud/kubernetes
    vars:
      command: create_deployment
      deployment_name: elasticsearch
      image: elasticsearch:8.5.0
      replicas: 1
      port: 9200

  - name: Expose Elasticsearch
    module: cloud/kubernetes
    vars:
      command: create_service
      service_name: elasticsearch
      service_type: ClusterIP
      port: 9200

  # 5. Deploy Fluentd for log collection
  - name: Deploy Fluentd DaemonSet
    module: cloud/kubernetes
    vars:
      command: apply_manifest
      manifest_content: |
        apiVersion: apps/v1
        kind: DaemonSet
        metadata:
          name: fluentd
          namespace: {{ namespace }}
        spec:
          selector:
            matchLabels:
              app: fluentd
          template:
            metadata:
              labels:
                app: fluentd
            spec:
              containers:
              - name: fluentd
                image: fluent/fluentd-kubernetes-daemonset:v1-debian-elasticsearch
                env:
                - name: FLUENT_ELASTICSEARCH_HOST
                  value: "elasticsearch.monitoring.svc.cluster.local"
                - name: FLUENT_ELASTICSEARCH_PORT
                  value: "9200"
                volumeMounts:
                - name: varlog
                  mountPath: /var/log
                - name: varlibdockercontainers
                  mountPath: /var/lib/docker/containers
                  readOnly: true
              volumes:
              - name: varlog
                hostPath:
                  path: /var/log
              - name: varlibdockercontainers
                hostPath:
                  path: /var/lib/docker/containers

  # 6. Deploy Kibana for log visualization
  - name: Deploy Kibana
    module: cloud/kubernetes
    vars:
      command: create_deployment
      deployment_name: kibana
      image: kibana:8.5.0
      replicas: 1
      port: 5601

  - name: Expose Kibana
    module: cloud/kubernetes
    vars:
      command: create_service
      service_name: kibana
      service_type: LoadBalancer
      port: 80
      target_port: 5601
```

---

## Best Practices Summary

1. **Always use namespaces** for logical isolation
2. **Use ConfigMaps and Secrets** for configuration management
3. **Implement health checks** (liveness and readiness probes)
4. **Set resource limits and requests** for all containers
5. **Use NetworkPolicies** to restrict traffic
6. **Implement RBAC** for access control
7. **Use ServiceAccounts** for pod identity
8. **Enable monitoring and logging** from the start
9. **Automate deployments** via CI/CD pipelines
10. **Plan for disaster recovery** with backups

## Additional Resources

- [Kubernetes Best Practices](https://kubernetes.io/docs/concepts/configuration/overview/)
- [12-Factor App Methodology](https://12factor.net/)
- [CNCF Cloud Native Trail Map](https://github.com/cncf/landscape/blob/master/README.md#trail-map)
