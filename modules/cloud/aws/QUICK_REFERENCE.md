# AWS Module - Quick Reference

## Common Commands

### EC2 Instances
```yaml
# Launch instance
command: create_instance
ami_id: ami-12345678
instance_type: t3.micro
key_name: my-key
subnet_id: subnet-12345678

# Stop instance
command: stop_instance
instance_id: i-12345678

# Start instance
command: start_instance
instance_id: i-12345678

# Terminate instance
command: terminate_instance
instance_id: i-12345678

# List instances
command: describe_instances
```

### S3 Storage
```yaml
# Create bucket
command: create_bucket
bucket_name: my-bucket

# Upload file
command: upload_object
bucket_name: my-bucket
local_file: /path/to/file
object_key: path/in/bucket

# Download file
command: download_object
bucket_name: my-bucket
object_key: path/in/bucket
local_file: /path/to/save

# List objects
command: list_objects
bucket_name: my-bucket
prefix: folder/

# Delete bucket
command: delete_bucket
bucket_name: my-bucket
```

### VPC and Networking
```yaml
# Create VPC
command: create_vpc
cidr_block: 10.0.0.0/16

# Create subnet
command: create_subnet
vpc_id: vpc-12345678
cidr_block: 10.0.1.0/24
availability_zone: us-east-1a

# Create security group
command: create_security_group
security_group_name: web-servers
security_group_description: Web server SG
vpc_id: vpc-12345678

# Allow HTTP
command: authorize_security_group_ingress
security_group_id: sg-12345678
ip_protocol: tcp
from_port: 80
to_port: 80
cidr_ip: 0.0.0.0/0
```

### IAM Management
```yaml
# Create user
command: create_user
user_name: john-doe

# Create role
command: create_role
role_name: ec2-role
assume_role_policy_document: |
  {
    "Version": "2012-10-17",
    "Statement": [{
      "Effect": "Allow",
      "Principal": {"Service": "ec2.amazonaws.com"},
      "Action": "sts:AssumeRole"
    }]
  }

# Attach policy
command: attach_role_policy
role_name: ec2-role
policy_arn: arn:aws:iam::aws:policy/ReadOnlyAccess

# Create access key
command: create_access_key
user_name: john-doe
```

### RDS Databases
```yaml
# Create database
command: create_db_instance
db_instance_identifier: mydb
db_instance_class: db.t3.micro
engine: mysql
master_username: admin
master_password: SecurePass123
allocated_storage: 20

# Create snapshot
command: create_db_snapshot
db_instance_identifier: mydb
snapshot_identifier: mydb-backup-2024-01-15

# Restore database
command: restore_db_instance
db_instance_identifier: mydb-restored
snapshot_identifier: mydb-backup-2024-01-15

# Delete database
command: delete_db_instance
db_instance_identifier: mydb
```

### Lambda Functions
```yaml
# Create function
command: create_function
function_name: my-function
runtime: python3.9
handler: index.handler
lambda_role: arn:aws:iam::123456789012:role/lambda-role
zip_file: /path/to/function.zip
environment_variables:
  DB_HOST: database.example.com
  API_KEY: secret

# Invoke function
command: invoke_function
function_name: my-function
payload: '{"key": "value"}'

# Delete function
command: delete_function
function_name: my-function
```

## Variable Reference

### Common Variables
```yaml
aws_region: us-east-1           # AWS region
aws_profile: default            # AWS CLI profile
aws_access_key_id: KEY          # Access key (optional)
aws_secret_access_key: SECRET   # Secret key (optional)
output_format: json             # Output format
```

### EC2 Variables
```yaml
instance_type: t3.micro         # Instance type
ami_id: ami-12345678            # AMI ID
instance_id: i-12345678         # Instance ID
key_name: my-key                # SSH key name
subnet_id: subnet-12345678      # Subnet ID
security_group_ids:             # Security groups
  - sg-12345678
```

### S3 Variables
```yaml
bucket_name: my-bucket          # Bucket name
object_key: path/to/file        # Object key
local_file: /local/path         # Local file path
prefix: folder/                 # Prefix for listing
bucket_policy: '{...}'          # Bucket policy JSON
```

### VPC Variables
```yaml
vpc_id: vpc-12345678            # VPC ID
cidr_block: 10.0.0.0/16         # CIDR block
subnet_id: subnet-12345678      # Subnet ID
availability_zone: us-east-1a   # AZ
security_group_id: sg-12345678  # Security group ID
security_group_name: web-sg     # SG name
ip_protocol: tcp                # Protocol
from_port: 80                   # Starting port
to_port: 80                     # Ending port
cidr_ip: 0.0.0.0/0             # CIDR IP
```

### IAM Variables
```yaml
user_name: john-doe             # User name
role_name: ec2-role             # Role name
policy_arn: arn:aws:iam::...    # Policy ARN
assume_role_policy_document:    # Trust policy
```

### RDS Variables
```yaml
db_instance_identifier: mydb    # DB instance ID
db_instance_class: db.t3.micro  # Instance class
engine: mysql                   # DB engine
master_username: admin          # Master user
master_password: pass           # Master password
allocated_storage: 20           # Storage GB
snapshot_identifier: snap-123   # Snapshot ID
```

### Lambda Variables
```yaml
function_name: my-function      # Function name
runtime: python3.9              # Runtime
handler: index.handler          # Handler
lambda_role: arn:aws:iam::...   # IAM role ARN
zip_file: /path/to/code.zip     # Code ZIP
payload: '{"key": "value"}'     # Invoke payload
environment_variables:          # Env vars
  VAR1: value1
  VAR2: value2
```

## Complete Stack Example

```yaml
name: Deploy Web Application
description: Complete web app with VPC, EC2, and RDS

inventory:
  hosts:
    - aws-management

defaults:
  aws_region: us-east-1
  environment: production

run:
  # 1. Create VPC
  - name: Create VPC
    module: cloud/aws
    vars:
      command: create_vpc
      cidr_block: 10.0.0.0/16

  # 2. Create subnet
  - name: Create public subnet
    module: cloud/aws
    vars:
      command: create_subnet
      vpc_id: "{{ facts.vpc_id }}"
      cidr_block: 10.0.1.0/24
      availability_zone: us-east-1a

  # 3. Create security group
  - name: Create security group
    module: cloud/aws
    vars:
      command: create_security_group
      security_group_name: web-servers
      vpc_id: "{{ facts.vpc_id }}"

  # 4. Allow HTTP
  - name: Allow HTTP traffic
    module: cloud/aws
    vars:
      command: authorize_security_group_ingress
      security_group_id: "{{ facts.sg_id }}"
      ip_protocol: tcp
      from_port: 80
      to_port: 80
      cidr_ip: 0.0.0.0/0

  # 5. Create database
  - name: Create RDS database
    module: cloud/aws
    vars:
      command: create_db_instance
      db_instance_identifier: webapp-db
      db_instance_class: db.t3.small
      engine: mysql
      master_username: admin
      master_password: "{{ vault.db_password }}"
      allocated_storage: 50

  # 6. Launch EC2
  - name: Launch web server
    module: cloud/aws
    vars:
      command: create_instance
      ami_id: ami-0c55b159cbfafe1f0
      instance_type: t3.medium
      key_name: production-key
      subnet_id: "{{ facts.subnet_id }}"
      security_group_ids:
        - "{{ facts.sg_id }}"

  # 7. Tag instance
  - name: Tag web server
    module: cloud/aws
    vars:
      command: create_tags
      instance_id: "{{ facts.instance_id }}"
      tag_specifications:
        Name: web-server-01
        Environment: "{{ var.environment }}"
        ManagedBy: OpenFroyo
```

## Authentication Methods

### Method 1: AWS Profile (Recommended)
```yaml
vars:
  aws_profile: production
  aws_region: us-east-1
```

### Method 2: Environment Variables
```yaml
# Set before running
export AWS_ACCESS_KEY_ID="..."
export AWS_SECRET_ACCESS_KEY="..."

vars:
  aws_region: us-east-1
```

### Method 3: Explicit Credentials
```yaml
vars:
  aws_access_key_id: AKIAIOSFODNN7EXAMPLE
  aws_secret_access_key: "{{ vault.aws_secret }}"
  aws_region: us-east-1
```

### Method 4: IAM Instance Role
```yaml
# No credentials needed on EC2 with IAM role
vars:
  aws_region: us-east-1
```

## Common Patterns

### Create and Tag Resource
```yaml
- name: Create instance
  module: cloud/aws
  vars:
    command: create_instance
    ami_id: ami-12345678

- name: Tag instance
  module: cloud/aws
  vars:
    command: create_tags
    instance_id: "{{ facts.instance_id }}"
    tag_specifications:
      Name: my-server
      Environment: prod
```

### VPC with Subnet and SG
```yaml
- name: Create VPC
  module: cloud/aws
  vars:
    command: create_vpc
    cidr_block: 10.0.0.0/16

- name: Create subnet
  module: cloud/aws
  vars:
    command: create_subnet
    vpc_id: "{{ facts.vpc_id }}"
    cidr_block: 10.0.1.0/24

- name: Create security group
  module: cloud/aws
  vars:
    command: create_security_group
    security_group_name: my-sg
    vpc_id: "{{ facts.vpc_id }}"
```

### S3 Upload Multiple Files
```yaml
- name: Upload HTML
  module: cloud/aws
  vars:
    command: upload_object
    bucket_name: my-bucket
    local_file: /var/www/index.html
    object_key: index.html

- name: Upload CSS
  module: cloud/aws
  vars:
    command: upload_object
    bucket_name: my-bucket
    local_file: /var/www/style.css
    object_key: css/style.css
```

## Tips and Best Practices

### Always Tag Resources
```yaml
tag_specifications:
  Name: descriptive-name
  Environment: production|staging|dev
  ManagedBy: OpenFroyo
  Owner: team-name
  CostCenter: department
```

### Use Variables for Reusability
```yaml
defaults:
  aws_region: us-east-1
  instance_type: t3.micro
  environment: production
```

### Secure Sensitive Data
```yaml
# Never hardcode passwords
master_password: "{{ vault.db_password }}"
api_key: "{{ vault.api_key }}"
```

### Reference Facts
```yaml
# Use outputs from previous tasks
vpc_id: "{{ facts.vpc_id }}"
instance_id: "{{ facts.instance_id }}"
```

### Multi-Region Deployments
```yaml
# Define regions
defaults:
  primary_region: us-east-1
  dr_region: us-west-2

# Use in tasks
run:
  - name: Create primary VPC
    vars:
      aws_region: "{{ var.primary_region }}"
```

## Error Handling

```yaml
# Continue on error
run:
  - name: Try to create instance
    module: cloud/aws
    vars:
      command: create_instance
    continue_on_error: true
```

## Parallel Execution

```yaml
# Execute on multiple hosts in parallel
inventory:
  hosts:
    - mgmt-01
    - mgmt-02

strategy: parallel
max_parallel: 2
```

## Documentation

- **[README.md](README.md)** - Complete user guide
- **[COMMANDS.md](COMMANDS.md)** - Detailed command reference
- **[WORKFLOWS.md](WORKFLOWS.md)** - Production workflows
- **[test.ofy](test.ofy)** - Test examples
