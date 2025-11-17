# AWS Module - Common Workflows

This document provides complete, production-ready workflows for common AWS deployment scenarios using OpenFroyo.

## Table of Contents
1. [VPC Setup from Scratch](#1-vpc-setup-from-scratch)
2. [Web Application with EC2 and RDS](#2-web-application-with-ec2-and-rds)
3. [Static Website with S3](#3-static-website-with-s3)
4. [Serverless API with Lambda](#4-serverless-api-with-lambda)
5. [Multi-Tier Application](#5-multi-tier-application)
6. [Disaster Recovery Setup](#6-disaster-recovery-setup)
7. [Development Environment Provisioning](#7-development-environment-provisioning)

---

## 1. VPC Setup from Scratch

Create a complete VPC with public and private subnets, internet gateway, and security groups.

**File:** `stacks/aws/setup_vpc.ofy`

```yaml
name: Complete VPC Setup
description: Create VPC with public/private subnets and security groups

inventory:
  hosts:
    - aws-management-01

defaults:
  aws_region: us-east-1
  vpc_cidr: 10.0.0.0/16
  public_subnet_cidr: 10.0.1.0/24
  private_subnet_cidr: 10.0.2.0/24

run:
  # Step 1: Create VPC
  - name: Create VPC
    module: cloud/aws
    vars:
      command: create_vpc
      cidr_block: "{{ var.vpc_cidr }}"
      aws_region: "{{ var.aws_region }}"

  # Step 2: Create Public Subnet
  - name: Create public subnet
    module: cloud/aws
    vars:
      command: create_subnet
      vpc_id: "{{ facts.vpc_id }}"
      cidr_block: "{{ var.public_subnet_cidr }}"
      availability_zone: us-east-1a
      aws_region: "{{ var.aws_region }}"

  # Step 3: Create Private Subnet
  - name: Create private subnet
    module: cloud/aws
    vars:
      command: create_subnet
      vpc_id: "{{ facts.vpc_id }}"
      cidr_block: "{{ var.private_subnet_cidr }}"
      availability_zone: us-east-1b
      aws_region: "{{ var.aws_region }}"

  # Step 4: Create Web Server Security Group
  - name: Create web server security group
    module: cloud/aws
    vars:
      command: create_security_group
      security_group_name: web-servers
      security_group_description: Security group for web servers
      vpc_id: "{{ facts.vpc_id }}"
      aws_region: "{{ var.aws_region }}"

  # Step 5: Allow HTTP Traffic
  - name: Allow HTTP traffic
    module: cloud/aws
    vars:
      command: authorize_security_group_ingress
      security_group_id: "{{ facts.security_group_id }}"
      ip_protocol: tcp
      from_port: 80
      to_port: 80
      cidr_ip: 0.0.0.0/0
      aws_region: "{{ var.aws_region }}"

  # Step 6: Allow HTTPS Traffic
  - name: Allow HTTPS traffic
    module: cloud/aws
    vars:
      command: authorize_security_group_ingress
      security_group_id: "{{ facts.security_group_id }}"
      ip_protocol: tcp
      from_port: 443
      to_port: 443
      cidr_ip: 0.0.0.0/0
      aws_region: "{{ var.aws_region }}"

  # Step 7: Allow SSH from Management Network
  - name: Allow SSH from management network
    module: cloud/aws
    vars:
      command: authorize_security_group_ingress
      security_group_id: "{{ facts.security_group_id }}"
      ip_protocol: tcp
      from_port: 22
      to_port: 22
      cidr_ip: 203.0.113.0/24  # Replace with your management network
      aws_region: "{{ var.aws_region }}"

  # Step 8: Create Database Security Group
  - name: Create database security group
    module: cloud/aws
    vars:
      command: create_security_group
      security_group_name: database-servers
      security_group_description: Security group for database servers
      vpc_id: "{{ facts.vpc_id }}"
      aws_region: "{{ var.aws_region }}"

  # Step 9: Allow MySQL from Web Servers
  - name: Allow MySQL from web servers
    module: cloud/aws
    vars:
      command: authorize_security_group_ingress
      security_group_id: "{{ facts.database_sg_id }}"
      ip_protocol: tcp
      from_port: 3306
      to_port: 3306
      cidr_ip: "{{ var.public_subnet_cidr }}"
      aws_region: "{{ var.aws_region }}"
```

**Run:**
```bash
froyo apply stacks/aws/setup_vpc.ofy
```

---

## 2. Web Application with EC2 and RDS

Deploy a complete web application with EC2 instances and RDS database.

**File:** `stacks/aws/deploy_webapp.ofy`

```yaml
name: Deploy Web Application
description: Deploy WordPress with EC2 and RDS MySQL

inventory:
  hosts:
    - aws-management-01

defaults:
  aws_region: us-east-1
  vpc_id: vpc-1234567890abcdef0  # From previous VPC setup
  subnet_id: subnet-1234567890abcdef0
  web_sg_id: sg-web-1234567890abcdef0
  db_sg_id: sg-db-1234567890abcdef0
  ami_id: ami-0c55b159cbfafe1f0  # Amazon Linux 2
  key_name: my-web-key

run:
  # Step 1: Create RDS MySQL Database
  - name: Create MySQL database for WordPress
    module: cloud/aws
    vars:
      command: create_db_instance
      db_instance_identifier: wordpress-db
      db_instance_class: db.t3.small
      engine: mysql
      master_username: wpdbadmin
      master_password: "{{ vault.db_password }}"  # Store in vault
      allocated_storage: 50
      aws_region: "{{ var.aws_region }}"

  # Step 2: Launch Web Server EC2 Instance
  - name: Launch WordPress web server
    module: cloud/aws
    vars:
      command: create_instance
      ami_id: "{{ var.ami_id }}"
      instance_type: t3.medium
      key_name: "{{ var.key_name }}"
      subnet_id: "{{ var.subnet_id }}"
      security_group_ids:
        - "{{ var.web_sg_id }}"
      aws_region: "{{ var.aws_region }}"

  # Step 3: Tag Web Server
  - name: Tag web server instance
    module: cloud/aws
    vars:
      command: create_tags
      instance_id: "{{ facts.instance_id }}"
      tag_specifications:
        Name: wordpress-web-01
        Application: WordPress
        Environment: production
        ManagedBy: OpenFroyo
      aws_region: "{{ var.aws_region }}"

  # Step 4: Create S3 Bucket for Media Files
  - name: Create S3 bucket for WordPress media
    module: cloud/aws
    vars:
      command: create_bucket
      bucket_name: wordpress-media-{{ var.environment }}-2024
      aws_region: "{{ var.aws_region }}"

  # Step 5: Wait for instance to be running (manual check for MVP)
  - name: Display instance information
    module: cloud/aws
    vars:
      command: describe_instances
      instance_id: "{{ facts.instance_id }}"
      aws_region: "{{ var.aws_region }}"

  # Step 6: Create AMI backup after initial setup
  - name: Create initial AMI backup
    module: cloud/aws
    vars:
      command: create_ami
      instance_id: "{{ facts.instance_id }}"
      image_name: wordpress-web-01-initial-{{ var.timestamp }}
      aws_region: "{{ var.aws_region }}"
```

**Post-Deployment Steps:**
After the stack completes, SSH into the instance and install WordPress:

```bash
# SSH to instance
ssh -i ~/.ssh/my-web-key.pem ec2-user@<instance-public-ip>

# Install WordPress
sudo yum update -y
sudo yum install -y httpd php php-mysqlnd
sudo systemctl start httpd
sudo systemctl enable httpd

# Download WordPress
cd /tmp
wget https://wordpress.org/latest.tar.gz
tar -xzf latest.tar.gz
sudo cp -r wordpress/* /var/www/html/

# Configure WordPress
sudo chown -R apache:apache /var/www/html/
```

---

## 3. Static Website with S3

Host a static website on S3 with public access.

**File:** `stacks/aws/static_website.ofy`

```yaml
name: Deploy Static Website to S3
description: Host static website on S3 with public read access

inventory:
  hosts:
    - aws-management-01

defaults:
  aws_region: us-west-2
  bucket_name: my-company-website-2024
  website_dir: /var/www/static-site

run:
  # Step 1: Create S3 Bucket
  - name: Create S3 bucket for website
    module: cloud/aws
    vars:
      command: create_bucket
      bucket_name: "{{ var.bucket_name }}"
      aws_region: "{{ var.aws_region }}"

  # Step 2: Upload index.html
  - name: Upload index page
    module: cloud/aws
    vars:
      command: upload_object
      bucket_name: "{{ var.bucket_name }}"
      local_file: "{{ var.website_dir }}/index.html"
      object_key: index.html
      aws_region: "{{ var.aws_region }}"

  # Step 3: Upload CSS files
  - name: Upload styles
    module: cloud/aws
    vars:
      command: upload_object
      bucket_name: "{{ var.bucket_name }}"
      local_file: "{{ var.website_dir }}/css/style.css"
      object_key: css/style.css
      aws_region: "{{ var.aws_region }}"

  # Step 4: Upload JavaScript files
  - name: Upload scripts
    module: cloud/aws
    vars:
      command: upload_object
      bucket_name: "{{ var.bucket_name }}"
      local_file: "{{ var.website_dir }}/js/app.js"
      object_key: js/app.js
      aws_region: "{{ var.aws_region }}"

  # Step 5: Upload images
  - name: Upload logo
    module: cloud/aws
    vars:
      command: upload_object
      bucket_name: "{{ var.bucket_name }}"
      local_file: "{{ var.website_dir }}/images/logo.png"
      object_key: images/logo.png
      aws_region: "{{ var.aws_region }}"

  # Step 6: Set public read policy
  - name: Make bucket publicly readable
    module: cloud/aws
    vars:
      command: set_bucket_policy
      bucket_name: "{{ var.bucket_name }}"
      bucket_policy: |
        {
          "Version": "2012-10-17",
          "Statement": [
            {
              "Sid": "PublicReadGetObject",
              "Effect": "Allow",
              "Principal": "*",
              "Action": "s3:GetObject",
              "Resource": "arn:aws:s3:::{{ var.bucket_name }}/*"
            }
          ]
        }
      aws_region: "{{ var.aws_region }}"
```

**Access Website:**
```
http://my-company-website-2024.s3-website-us-west-2.amazonaws.com
```

---

## 4. Serverless API with Lambda

Create a serverless API using Lambda functions.

**File:** `stacks/aws/serverless_api.ofy`

```yaml
name: Deploy Serverless API
description: Create Lambda functions for REST API

inventory:
  hosts:
    - aws-management-01

defaults:
  aws_region: us-east-1
  lambda_role_arn: arn:aws:iam::123456789012:role/lambda-execution-role
  functions_dir: /opt/lambda-functions

run:
  # Step 1: Create Lambda for User Registration
  - name: Create user registration function
    module: cloud/aws
    vars:
      command: create_function
      function_name: api-user-register
      runtime: python3.9
      handler: register.lambda_handler
      lambda_role: "{{ var.lambda_role_arn }}"
      zip_file: "{{ var.functions_dir }}/register.zip"
      environment_variables:
        DB_HOST: database.example.com
        DB_NAME: users
        ENVIRONMENT: production
      aws_region: "{{ var.aws_region }}"

  # Step 2: Create Lambda for User Login
  - name: Create user login function
    module: cloud/aws
    vars:
      command: create_function
      function_name: api-user-login
      runtime: python3.9
      handler: login.lambda_handler
      lambda_role: "{{ var.lambda_role_arn }}"
      zip_file: "{{ var.functions_dir }}/login.zip"
      environment_variables:
        DB_HOST: database.example.com
        DB_NAME: users
        JWT_SECRET: "{{ vault.jwt_secret }}"
        ENVIRONMENT: production
      aws_region: "{{ var.aws_region }}"

  # Step 3: Create Lambda for Data Processing
  - name: Create data processing function
    module: cloud/aws
    vars:
      command: create_function
      function_name: api-process-data
      runtime: python3.9
      handler: process.lambda_handler
      lambda_role: "{{ var.lambda_role_arn }}"
      zip_file: "{{ var.functions_dir }}/process.zip"
      environment_variables:
        S3_BUCKET: data-processing-results
        SQS_QUEUE: https://sqs.us-east-1.amazonaws.com/123456789012/processing-queue
        ENVIRONMENT: production
      aws_region: "{{ var.aws_region }}"

  # Step 4: Test registration function
  - name: Test user registration endpoint
    module: cloud/aws
    vars:
      command: invoke_function
      function_name: api-user-register
      payload: '{"email": "test@example.com", "name": "Test User"}'
      aws_region: "{{ var.aws_region }}"
```

**Lambda Function Example (register.py):**
```python
import json
import os

def lambda_handler(event, context):
    db_host = os.environ['DB_HOST']
    db_name = os.environ['DB_NAME']

    # Parse request
    body = json.loads(event.get('body', '{}'))
    email = body.get('email')
    name = body.get('name')

    # Register user (implement your logic)
    # ...

    return {
        'statusCode': 200,
        'body': json.dumps({
            'message': 'User registered successfully',
            'user_id': '12345'
        })
    }
```

---

## 5. Multi-Tier Application

Deploy a complete multi-tier application with load balancer, application servers, and database.

**File:** `stacks/aws/multi_tier_app.ofy`

```yaml
name: Multi-Tier Application Deployment
description: Deploy application with web tier, app tier, and database tier

inventory:
  hosts:
    - aws-management-01

defaults:
  aws_region: us-east-1
  vpc_id: vpc-1234567890abcdef0
  public_subnet_id: subnet-public-123
  private_subnet_id: subnet-private-456
  web_ami: ami-web-123
  app_ami: ami-app-456
  key_name: production-key

run:
  # Database Tier
  - name: Create RDS database
    module: cloud/aws
    vars:
      command: create_db_instance
      db_instance_identifier: app-production-db
      db_instance_class: db.t3.medium
      engine: postgres
      master_username: dbadmin
      master_password: "{{ vault.db_password }}"
      allocated_storage: 100
      aws_region: "{{ var.aws_region }}"

  # Application Tier - Server 1
  - name: Launch application server 1
    module: cloud/aws
    vars:
      command: create_instance
      ami_id: "{{ var.app_ami }}"
      instance_type: t3.large
      key_name: "{{ var.key_name }}"
      subnet_id: "{{ var.private_subnet_id }}"
      security_group_ids:
        - sg-app-servers
      aws_region: "{{ var.aws_region }}"

  - name: Tag application server 1
    module: cloud/aws
    vars:
      command: create_tags
      instance_id: "{{ facts.app_instance_1_id }}"
      tag_specifications:
        Name: app-server-01
        Tier: application
        Environment: production
      aws_region: "{{ var.aws_region }}"

  # Application Tier - Server 2
  - name: Launch application server 2
    module: cloud/aws
    vars:
      command: create_instance
      ami_id: "{{ var.app_ami }}"
      instance_type: t3.large
      key_name: "{{ var.key_name }}"
      subnet_id: "{{ var.private_subnet_id }}"
      security_group_ids:
        - sg-app-servers
      aws_region: "{{ var.aws_region }}"

  - name: Tag application server 2
    module: cloud/aws
    vars:
      command: create_tags
      instance_id: "{{ facts.app_instance_2_id }}"
      tag_specifications:
        Name: app-server-02
        Tier: application
        Environment: production
      aws_region: "{{ var.aws_region }}"

  # Web Tier - Server 1
  - name: Launch web server 1
    module: cloud/aws
    vars:
      command: create_instance
      ami_id: "{{ var.web_ami }}"
      instance_type: t3.medium
      key_name: "{{ var.key_name }}"
      subnet_id: "{{ var.public_subnet_id }}"
      security_group_ids:
        - sg-web-servers
      aws_region: "{{ var.aws_region }}"

  - name: Tag web server 1
    module: cloud/aws
    vars:
      command: create_tags
      instance_id: "{{ facts.web_instance_1_id }}"
      tag_specifications:
        Name: web-server-01
        Tier: web
        Environment: production
      aws_region: "{{ var.aws_region }}"

  # Web Tier - Server 2
  - name: Launch web server 2
    module: cloud/aws
    vars:
      command: create_instance
      ami_id: "{{ var.web_ami }}"
      instance_type: t3.medium
      key_name: "{{ var.key_name }}"
      subnet_id: "{{ var.public_subnet_id }}"
      security_group_ids:
        - sg-web-servers
      aws_region: "{{ var.aws_region }}"

  - name: Tag web server 2
    module: cloud/aws
    vars:
      command: create_tags
      instance_id: "{{ facts.web_instance_2_id }}"
      tag_specifications:
        Name: web-server-02
        Tier: web
        Environment: production
      aws_region: "{{ var.aws_region }}"

  # Verify deployment
  - name: List all instances
    module: cloud/aws
    vars:
      command: describe_instances
      aws_region: "{{ var.aws_region }}"
```

---

## 6. Disaster Recovery Setup

Create backups and disaster recovery resources.

**File:** `stacks/aws/disaster_recovery.ofy`

```yaml
name: Disaster Recovery Setup
description: Create snapshots and backups for disaster recovery

inventory:
  hosts:
    - aws-management-01

defaults:
  aws_region: us-east-1
  dr_region: us-west-2
  timestamp: "{{ now | date('Y-m-d-H-i-s') }}"

run:
  # Primary Region Backups
  - name: Create RDS snapshot
    module: cloud/aws
    vars:
      command: create_db_snapshot
      db_instance_identifier: production-db
      snapshot_identifier: production-db-backup-{{ var.timestamp }}
      aws_region: "{{ var.aws_region }}"

  - name: Create AMI of web server
    module: cloud/aws
    vars:
      command: create_ami
      instance_id: i-web-server-01
      image_name: web-server-backup-{{ var.timestamp }}
      aws_region: "{{ var.aws_region }}"

  - name: Create AMI of app server
    module: cloud/aws
    vars:
      command: create_ami
      instance_id: i-app-server-01
      image_name: app-server-backup-{{ var.timestamp }}
      aws_region: "{{ var.aws_region }}"

  # Backup application data to S3
  - name: Upload application logs to S3
    module: cloud/aws
    vars:
      command: upload_object
      bucket_name: disaster-recovery-backups
      local_file: /var/log/application/app.log
      object_key: backups/{{ var.timestamp }}/app.log
      aws_region: "{{ var.aws_region }}"

  - name: Upload database dump to S3
    module: cloud/aws
    vars:
      command: upload_object
      bucket_name: disaster-recovery-backups
      local_file: /backup/database-dump.sql.gz
      object_key: backups/{{ var.timestamp }}/database-dump.sql.gz
      aws_region: "{{ var.aws_region }}"

  # DR Region - Create resources
  - name: Create DR VPC in secondary region
    module: cloud/aws
    vars:
      command: create_vpc
      cidr_block: 10.1.0.0/16
      aws_region: "{{ var.dr_region }}"

  - name: Create DR subnet
    module: cloud/aws
    vars:
      command: create_subnet
      vpc_id: "{{ facts.dr_vpc_id }}"
      cidr_block: 10.1.1.0/24
      aws_region: "{{ var.dr_region }}"
```

---

## 7. Development Environment Provisioning

Quickly provision development environments for teams.

**File:** `stacks/aws/dev_environment.ofy`

```yaml
name: Development Environment Provisioning
description: Create isolated development environment for a developer

inventory:
  hosts:
    - aws-management-01

defaults:
  aws_region: us-east-1
  developer_name: john-doe
  vpc_id: vpc-dev-shared
  subnet_id: subnet-dev-shared

run:
  # Create developer's EC2 instance
  - name: Launch development instance
    module: cloud/aws
    vars:
      command: create_instance
      ami_id: ami-dev-baseline
      instance_type: t3.large
      key_name: dev-team-key
      subnet_id: "{{ var.subnet_id }}"
      security_group_ids:
        - sg-dev-instances
      aws_region: "{{ var.aws_region }}"

  - name: Tag development instance
    module: cloud/aws
    vars:
      command: create_tags
      instance_id: "{{ facts.instance_id }}"
      tag_specifications:
        Name: dev-{{ var.developer_name }}
        Owner: "{{ var.developer_name }}"
        Environment: development
        AutoShutdown: "true"
        ShutdownTime: "19:00"
      aws_region: "{{ var.aws_region }}"

  # Create developer's S3 bucket
  - name: Create developer S3 bucket
    module: cloud/aws
    vars:
      command: create_bucket
      bucket_name: dev-{{ var.developer_name }}-workspace
      aws_region: "{{ var.aws_region }}"

  # Create developer's RDS instance
  - name: Create development database
    module: cloud/aws
    vars:
      command: create_db_instance
      db_instance_identifier: dev-db-{{ var.developer_name }}
      db_instance_class: db.t3.micro
      engine: mysql
      master_username: devuser
      master_password: "{{ vault.dev_db_password }}"
      allocated_storage: 20
      aws_region: "{{ var.aws_region }}"

  # Create IAM user for developer
  - name: Create IAM user
    module: cloud/aws
    vars:
      command: create_user
      user_name: dev-{{ var.developer_name }}

  - name: Create access keys
    module: cloud/aws
    vars:
      command: create_access_key
      user_name: dev-{{ var.developer_name }}
```

---

## Best Practices for Workflows

### 1. Use Variables for Reusability
Define defaults at the top of your stack files:
```yaml
defaults:
  aws_region: us-east-1
  environment: production
  managed_by: OpenFroyo
```

### 2. Tag All Resources
Always tag resources for cost tracking and management:
```yaml
tag_specifications:
  Name: resource-name
  Environment: "{{ var.environment }}"
  ManagedBy: OpenFroyo
  CostCenter: engineering
  Owner: team-name
```

### 3. Use Fact References
Chain tasks together using facts from previous commands:
```yaml
vpc_id: "{{ facts.vpc_id }}"
instance_id: "{{ facts.instance_id }}"
```

### 4. Secure Sensitive Data
Never hardcode passwords or keys:
```yaml
master_password: "{{ vault.db_password }}"
api_key: "{{ vault.api_key }}"
```

### 5. Document Your Stacks
Add clear names and descriptions:
```yaml
name: Production Database Deployment
description: Deploy production PostgreSQL RDS instance with read replicas
```

### 6. Plan for Cleanup
Create corresponding cleanup stacks:
```yaml
# File: stacks/aws/cleanup_dev_environment.ofy
run:
  - name: Terminate instance
    module: cloud/aws
    vars:
      command: terminate_instance
      instance_id: "{{ var.instance_id }}"
```

### 7. Test in Non-Production First
Always test stacks in development before running in production:
```bash
# Test in dev
froyo apply stacks/aws/deploy_webapp.ofy --extra-vars "environment=dev"

# Deploy to production
froyo apply stacks/aws/deploy_webapp.ofy --extra-vars "environment=production"
```
