# AWS Module - Command Reference

Complete reference for all AWS module commands.

## Table of Contents
- [EC2 Instance Management](#ec2-instance-management)
- [S3 Storage Management](#s3-storage-management)
- [VPC and Networking](#vpc-and-networking)
- [IAM Management](#iam-management)
- [RDS Database Management](#rds-database-management)
- [Lambda Functions](#lambda-functions)

---

## EC2 Instance Management

### create_instance

Launch a new EC2 instance.

**Required Variables:**
- `ami_id` - AMI ID to launch

**Optional Variables:**
- `instance_type` - Instance type (default: t3.micro)
- `key_name` - SSH key pair name
- `subnet_id` - VPC subnet ID
- `security_group_ids` - List of security group IDs

**Example:**
```yaml
- name: Launch web server
  module: cloud/aws
  vars:
    command: create_instance
    ami_id: ami-0c55b159cbfafe1f0
    instance_type: t3.medium
    key_name: web-server-key
    subnet_id: subnet-abc123
    security_group_ids:
      - sg-web-servers
    aws_region: us-east-1
```

**AWS CLI Equivalent:**
```bash
aws ec2 run-instances \
  --image-id ami-0c55b159cbfafe1f0 \
  --instance-type t3.medium \
  --key-name web-server-key \
  --subnet-id subnet-abc123 \
  --security-group-ids sg-web-servers \
  --region us-east-1
```

---

### terminate_instance

Terminate an EC2 instance.

**Required Variables:**
- `instance_id` - Instance ID to terminate

**Example:**
```yaml
- name: Terminate instance
  module: cloud/aws
  vars:
    command: terminate_instance
    instance_id: i-1234567890abcdef0
    aws_region: us-east-1
```

**AWS CLI Equivalent:**
```bash
aws ec2 terminate-instances --instance-ids i-1234567890abcdef0 --region us-east-1
```

---

### start_instance

Start a stopped EC2 instance.

**Required Variables:**
- `instance_id` - Instance ID to start

**Example:**
```yaml
- name: Start instance
  module: cloud/aws
  vars:
    command: start_instance
    instance_id: i-1234567890abcdef0
    aws_region: us-east-1
```

**AWS CLI Equivalent:**
```bash
aws ec2 start-instances --instance-ids i-1234567890abcdef0 --region us-east-1
```

---

### stop_instance

Stop a running EC2 instance.

**Required Variables:**
- `instance_id` - Instance ID to stop

**Example:**
```yaml
- name: Stop instance
  module: cloud/aws
  vars:
    command: stop_instance
    instance_id: i-1234567890abcdef0
    aws_region: us-east-1
```

**AWS CLI Equivalent:**
```bash
aws ec2 stop-instances --instance-ids i-1234567890abcdef0 --region us-east-1
```

---

### reboot_instance

Reboot an EC2 instance.

**Required Variables:**
- `instance_id` - Instance ID to reboot

**Example:**
```yaml
- name: Reboot instance
  module: cloud/aws
  vars:
    command: reboot_instance
    instance_id: i-1234567890abcdef0
    aws_region: us-east-1
```

**AWS CLI Equivalent:**
```bash
aws ec2 reboot-instances --instance-ids i-1234567890abcdef0 --region us-east-1
```

---

### describe_instances

Get information about EC2 instances.

**Optional Variables:**
- `instance_id` - Specific instance ID to describe

**Example:**
```yaml
- name: List all instances
  module: cloud/aws
  vars:
    command: describe_instances
    aws_region: us-east-1

- name: Get specific instance
  module: cloud/aws
  vars:
    command: describe_instances
    instance_id: i-1234567890abcdef0
    aws_region: us-east-1
```

**AWS CLI Equivalent:**
```bash
aws ec2 describe-instances --region us-east-1
aws ec2 describe-instances --instance-ids i-1234567890abcdef0 --region us-east-1
```

---

### create_ami

Create an AMI from an EC2 instance.

**Required Variables:**
- `instance_id` - Source instance ID
- `image_name` - Name for the new AMI

**Example:**
```yaml
- name: Create AMI backup
  module: cloud/aws
  vars:
    command: create_ami
    instance_id: i-1234567890abcdef0
    image_name: web-server-backup-2024-01-15
    aws_region: us-east-1
```

**AWS CLI Equivalent:**
```bash
aws ec2 create-image \
  --instance-id i-1234567890abcdef0 \
  --name web-server-backup-2024-01-15 \
  --region us-east-1
```

---

### describe_amis

List available AMIs owned by your account.

**Example:**
```yaml
- name: List my AMIs
  module: cloud/aws
  vars:
    command: describe_amis
    aws_region: us-east-1
```

**AWS CLI Equivalent:**
```bash
aws ec2 describe-images --owners self --region us-east-1
```

---

### create_tags

Add tags to EC2 resources.

**Required Variables:**
- `instance_id` - Resource ID to tag
- `tag_specifications` - Object with key-value pairs

**Example:**
```yaml
- name: Tag instance
  module: cloud/aws
  vars:
    command: create_tags
    instance_id: i-1234567890abcdef0
    tag_specifications:
      Name: web-server-01
      Environment: production
      Application: wordpress
    aws_region: us-east-1
```

**AWS CLI Equivalent:**
```bash
aws ec2 create-tags \
  --resources i-1234567890abcdef0 \
  --tags Key=Name,Value=web-server-01 Key=Environment,Value=production \
  --region us-east-1
```

---

### describe_instance_types

List available EC2 instance types.

**Example:**
```yaml
- name: List instance types
  module: cloud/aws
  vars:
    command: describe_instance_types
    aws_region: us-east-1
```

**AWS CLI Equivalent:**
```bash
aws ec2 describe-instance-types --region us-east-1
```

---

## S3 Storage Management

### create_bucket

Create a new S3 bucket.

**Required Variables:**
- `bucket_name` - Name for the bucket (must be globally unique)

**Example:**
```yaml
- name: Create S3 bucket
  module: cloud/aws
  vars:
    command: create_bucket
    bucket_name: my-application-assets-2024
    aws_region: us-west-2
```

**AWS CLI Equivalent:**
```bash
aws s3 mb s3://my-application-assets-2024 --region us-west-2
```

---

### delete_bucket

Delete an S3 bucket (including all objects).

**Required Variables:**
- `bucket_name` - Name of bucket to delete

**Example:**
```yaml
- name: Delete S3 bucket
  module: cloud/aws
  vars:
    command: delete_bucket
    bucket_name: my-old-bucket
    aws_region: us-west-2
```

**AWS CLI Equivalent:**
```bash
aws s3 rb s3://my-old-bucket --force
```

---

### list_buckets

List all S3 buckets in your account.

**Example:**
```yaml
- name: List all buckets
  module: cloud/aws
  vars:
    command: list_buckets
```

**AWS CLI Equivalent:**
```bash
aws s3 ls
```

---

### upload_object

Upload a file to S3.

**Required Variables:**
- `bucket_name` - Destination bucket
- `local_file` - Path to local file
- `object_key` - S3 object key (path in bucket)

**Example:**
```yaml
- name: Upload website files
  module: cloud/aws
  vars:
    command: upload_object
    bucket_name: my-website-assets
    local_file: /var/www/html/index.html
    object_key: public/index.html
    aws_region: us-west-2
```

**AWS CLI Equivalent:**
```bash
aws s3 cp /var/www/html/index.html s3://my-website-assets/public/index.html --region us-west-2
```

---

### download_object

Download a file from S3.

**Required Variables:**
- `bucket_name` - Source bucket
- `object_key` - S3 object key
- `local_file` - Destination local path

**Example:**
```yaml
- name: Download backup
  module: cloud/aws
  vars:
    command: download_object
    bucket_name: my-backups
    object_key: database/backup-2024-01-15.sql.gz
    local_file: /tmp/backup.sql.gz
    aws_region: us-west-2
```

**AWS CLI Equivalent:**
```bash
aws s3 cp s3://my-backups/database/backup-2024-01-15.sql.gz /tmp/backup.sql.gz --region us-west-2
```

---

### delete_object

Delete an object from S3.

**Required Variables:**
- `bucket_name` - Bucket containing object
- `object_key` - Object key to delete

**Example:**
```yaml
- name: Delete old file
  module: cloud/aws
  vars:
    command: delete_object
    bucket_name: my-bucket
    object_key: old/file.txt
    aws_region: us-west-2
```

**AWS CLI Equivalent:**
```bash
aws s3 rm s3://my-bucket/old/file.txt --region us-west-2
```

---

### list_objects

List objects in an S3 bucket.

**Required Variables:**
- `bucket_name` - Bucket to list

**Optional Variables:**
- `prefix` - Filter by prefix

**Example:**
```yaml
- name: List all objects
  module: cloud/aws
  vars:
    command: list_objects
    bucket_name: my-bucket
    aws_region: us-west-2

- name: List with prefix
  module: cloud/aws
  vars:
    command: list_objects
    bucket_name: my-bucket
    prefix: backups/2024/
    aws_region: us-west-2
```

**AWS CLI Equivalent:**
```bash
aws s3 ls s3://my-bucket --recursive
aws s3 ls s3://my-bucket/backups/2024/ --recursive
```

---

### set_bucket_policy

Set an S3 bucket policy.

**Required Variables:**
- `bucket_name` - Bucket name
- `bucket_policy` - JSON policy document

**Example:**
```yaml
- name: Set public read policy
  module: cloud/aws
  vars:
    command: set_bucket_policy
    bucket_name: my-public-bucket
    bucket_policy: |
      {
        "Version": "2012-10-17",
        "Statement": [{
          "Effect": "Allow",
          "Principal": "*",
          "Action": "s3:GetObject",
          "Resource": "arn:aws:s3:::my-public-bucket/*"
        }]
      }
    aws_region: us-west-2
```

**AWS CLI Equivalent:**
```bash
aws s3api put-bucket-policy --bucket my-public-bucket --policy file://policy.json --region us-west-2
```

---

## VPC and Networking

### create_vpc

Create a new VPC.

**Required Variables:**
- `cidr_block` - CIDR block for the VPC

**Example:**
```yaml
- name: Create VPC
  module: cloud/aws
  vars:
    command: create_vpc
    cidr_block: 10.0.0.0/16
    aws_region: us-east-1
```

**AWS CLI Equivalent:**
```bash
aws ec2 create-vpc --cidr-block 10.0.0.0/16 --region us-east-1
```

---

### delete_vpc

Delete a VPC.

**Required Variables:**
- `vpc_id` - VPC ID to delete

**Example:**
```yaml
- name: Delete VPC
  module: cloud/aws
  vars:
    command: delete_vpc
    vpc_id: vpc-1234567890abcdef0
    aws_region: us-east-1
```

**AWS CLI Equivalent:**
```bash
aws ec2 delete-vpc --vpc-id vpc-1234567890abcdef0 --region us-east-1
```

---

### create_subnet

Create a subnet in a VPC.

**Required Variables:**
- `vpc_id` - Parent VPC ID
- `cidr_block` - CIDR block for subnet

**Optional Variables:**
- `availability_zone` - Availability zone

**Example:**
```yaml
- name: Create public subnet
  module: cloud/aws
  vars:
    command: create_subnet
    vpc_id: vpc-1234567890abcdef0
    cidr_block: 10.0.1.0/24
    availability_zone: us-east-1a
    aws_region: us-east-1
```

**AWS CLI Equivalent:**
```bash
aws ec2 create-subnet \
  --vpc-id vpc-1234567890abcdef0 \
  --cidr-block 10.0.1.0/24 \
  --availability-zone us-east-1a \
  --region us-east-1
```

---

### delete_subnet

Delete a subnet.

**Required Variables:**
- `subnet_id` - Subnet ID to delete

**Example:**
```yaml
- name: Delete subnet
  module: cloud/aws
  vars:
    command: delete_subnet
    subnet_id: subnet-1234567890abcdef0
    aws_region: us-east-1
```

**AWS CLI Equivalent:**
```bash
aws ec2 delete-subnet --subnet-id subnet-1234567890abcdef0 --region us-east-1
```

---

### create_security_group

Create a security group.

**Required Variables:**
- `security_group_name` - Name for the security group

**Optional Variables:**
- `security_group_description` - Description (defaults to name)
- `vpc_id` - VPC ID (for VPC security groups)

**Example:**
```yaml
- name: Create web server security group
  module: cloud/aws
  vars:
    command: create_security_group
    security_group_name: web-servers
    security_group_description: Security group for web servers
    vpc_id: vpc-1234567890abcdef0
    aws_region: us-east-1
```

**AWS CLI Equivalent:**
```bash
aws ec2 create-security-group \
  --group-name web-servers \
  --description "Security group for web servers" \
  --vpc-id vpc-1234567890abcdef0 \
  --region us-east-1
```

---

### delete_security_group

Delete a security group.

**Required Variables:**
- `security_group_id` - Security group ID to delete

**Example:**
```yaml
- name: Delete security group
  module: cloud/aws
  vars:
    command: delete_security_group
    security_group_id: sg-1234567890abcdef0
    aws_region: us-east-1
```

**AWS CLI Equivalent:**
```bash
aws ec2 delete-security-group --group-id sg-1234567890abcdef0 --region us-east-1
```

---

### authorize_security_group_ingress

Add an ingress rule to a security group.

**Required Variables:**
- `security_group_id` - Security group ID
- `from_port` - Starting port
- `to_port` - Ending port

**Optional Variables:**
- `ip_protocol` - Protocol (default: tcp)
- `cidr_ip` - CIDR IP range (default: 0.0.0.0/0)

**Example:**
```yaml
- name: Allow HTTP traffic
  module: cloud/aws
  vars:
    command: authorize_security_group_ingress
    security_group_id: sg-1234567890abcdef0
    ip_protocol: tcp
    from_port: 80
    to_port: 80
    cidr_ip: 0.0.0.0/0
    aws_region: us-east-1

- name: Allow HTTPS from specific IP
  module: cloud/aws
  vars:
    command: authorize_security_group_ingress
    security_group_id: sg-1234567890abcdef0
    ip_protocol: tcp
    from_port: 443
    to_port: 443
    cidr_ip: 203.0.113.0/24
    aws_region: us-east-1
```

**AWS CLI Equivalent:**
```bash
aws ec2 authorize-security-group-ingress \
  --group-id sg-1234567890abcdef0 \
  --protocol tcp \
  --port 80-80 \
  --cidr 0.0.0.0/0 \
  --region us-east-1
```

---

### describe_security_groups

List security groups.

**Optional Variables:**
- `security_group_id` - Specific security group to describe

**Example:**
```yaml
- name: List all security groups
  module: cloud/aws
  vars:
    command: describe_security_groups
    aws_region: us-east-1
```

**AWS CLI Equivalent:**
```bash
aws ec2 describe-security-groups --region us-east-1
```

---

## IAM Management

### create_user

Create an IAM user.

**Required Variables:**
- `user_name` - Username for the new user

**Example:**
```yaml
- name: Create IAM user
  module: cloud/aws
  vars:
    command: create_user
    user_name: developer-john
```

**AWS CLI Equivalent:**
```bash
aws iam create-user --user-name developer-john
```

---

### delete_user

Delete an IAM user.

**Required Variables:**
- `user_name` - Username to delete

**Example:**
```yaml
- name: Delete IAM user
  module: cloud/aws
  vars:
    command: delete_user
    user_name: old-developer
```

**AWS CLI Equivalent:**
```bash
aws iam delete-user --user-name old-developer
```

---

### create_role

Create an IAM role.

**Required Variables:**
- `role_name` - Name for the role
- `assume_role_policy_document` - Trust policy JSON or file path

**Example:**
```yaml
- name: Create EC2 role
  module: cloud/aws
  vars:
    command: create_role
    role_name: ec2-web-server-role
    assume_role_policy_document: |
      {
        "Version": "2012-10-17",
        "Statement": [{
          "Effect": "Allow",
          "Principal": {"Service": "ec2.amazonaws.com"},
          "Action": "sts:AssumeRole"
        }]
      }
```

**AWS CLI Equivalent:**
```bash
aws iam create-role \
  --role-name ec2-web-server-role \
  --assume-role-policy-document file://trust-policy.json
```

---

### delete_role

Delete an IAM role.

**Required Variables:**
- `role_name` - Role name to delete

**Example:**
```yaml
- name: Delete role
  module: cloud/aws
  vars:
    command: delete_role
    role_name: old-role
```

**AWS CLI Equivalent:**
```bash
aws iam delete-role --role-name old-role
```

---

### attach_role_policy

Attach a managed policy to a role.

**Required Variables:**
- `role_name` - Role name
- `policy_arn` - Policy ARN to attach

**Example:**
```yaml
- name: Attach S3 read-only policy
  module: cloud/aws
  vars:
    command: attach_role_policy
    role_name: ec2-web-server-role
    policy_arn: arn:aws:iam::aws:policy/AmazonS3ReadOnlyAccess
```

**AWS CLI Equivalent:**
```bash
aws iam attach-role-policy \
  --role-name ec2-web-server-role \
  --policy-arn arn:aws:iam::aws:policy/AmazonS3ReadOnlyAccess
```

---

### create_access_key

Create an access key for an IAM user.

**Required Variables:**
- `user_name` - User to create access key for

**Example:**
```yaml
- name: Create access key
  module: cloud/aws
  vars:
    command: create_access_key
    user_name: developer-john
```

**AWS CLI Equivalent:**
```bash
aws iam create-access-key --user-name developer-john
```

---

## RDS Database Management

### create_db_instance

Create an RDS database instance.

**Required Variables:**
- `db_instance_identifier` - Database instance identifier
- `master_username` - Master username
- `master_password` - Master password

**Optional Variables:**
- `db_instance_class` - Instance class (default: db.t3.micro)
- `engine` - Database engine (default: mysql)
- `allocated_storage` - Storage in GB (default: 20)

**Example:**
```yaml
- name: Create MySQL database
  module: cloud/aws
  vars:
    command: create_db_instance
    db_instance_identifier: wordpress-db
    db_instance_class: db.t3.small
    engine: mysql
    master_username: admin
    master_password: SecurePassword123!
    allocated_storage: 50
    aws_region: us-east-1
```

**AWS CLI Equivalent:**
```bash
aws rds create-db-instance \
  --db-instance-identifier wordpress-db \
  --db-instance-class db.t3.small \
  --engine mysql \
  --master-username admin \
  --master-user-password SecurePassword123! \
  --allocated-storage 50 \
  --region us-east-1
```

---

### delete_db_instance

Delete an RDS database instance.

**Required Variables:**
- `db_instance_identifier` - Database instance to delete

**Example:**
```yaml
- name: Delete database
  module: cloud/aws
  vars:
    command: delete_db_instance
    db_instance_identifier: old-database
    aws_region: us-east-1
```

**AWS CLI Equivalent:**
```bash
aws rds delete-db-instance \
  --db-instance-identifier old-database \
  --skip-final-snapshot \
  --region us-east-1
```

---

### describe_db_instances

List RDS database instances.

**Optional Variables:**
- `db_instance_identifier` - Specific database to describe

**Example:**
```yaml
- name: List all databases
  module: cloud/aws
  vars:
    command: describe_db_instances
    aws_region: us-east-1
```

**AWS CLI Equivalent:**
```bash
aws rds describe-db-instances --region us-east-1
```

---

### create_db_snapshot

Create a snapshot of an RDS database.

**Required Variables:**
- `db_instance_identifier` - Source database
- `snapshot_identifier` - Name for the snapshot

**Example:**
```yaml
- name: Create database backup
  module: cloud/aws
  vars:
    command: create_db_snapshot
    db_instance_identifier: wordpress-db
    snapshot_identifier: wordpress-db-backup-2024-01-15
    aws_region: us-east-1
```

**AWS CLI Equivalent:**
```bash
aws rds create-db-snapshot \
  --db-instance-identifier wordpress-db \
  --db-snapshot-identifier wordpress-db-backup-2024-01-15 \
  --region us-east-1
```

---

### restore_db_instance

Restore a database from a snapshot.

**Required Variables:**
- `db_instance_identifier` - Name for restored database
- `snapshot_identifier` - Snapshot to restore from

**Example:**
```yaml
- name: Restore from backup
  module: cloud/aws
  vars:
    command: restore_db_instance
    db_instance_identifier: wordpress-db-restored
    snapshot_identifier: wordpress-db-backup-2024-01-15
    aws_region: us-east-1
```

**AWS CLI Equivalent:**
```bash
aws rds restore-db-instance-from-db-snapshot \
  --db-instance-identifier wordpress-db-restored \
  --db-snapshot-identifier wordpress-db-backup-2024-01-15 \
  --region us-east-1
```

---

## Lambda Functions

### create_function

Create a Lambda function.

**Required Variables:**
- `function_name` - Function name
- `lambda_role` - IAM role ARN for the function
- `zip_file` - Path to deployment package ZIP

**Optional Variables:**
- `runtime` - Runtime (default: python3.9)
- `handler` - Handler (default: index.handler)
- `environment_variables` - Environment variables object

**Example:**
```yaml
- name: Create Lambda function
  module: cloud/aws
  vars:
    command: create_function
    function_name: process-orders
    runtime: python3.9
    handler: lambda_function.lambda_handler
    lambda_role: arn:aws:iam::123456789012:role/lambda-execution-role
    zip_file: /path/to/function.zip
    environment_variables:
      DB_HOST: database.example.com
      API_KEY: secret-key
    aws_region: us-east-1
```

**AWS CLI Equivalent:**
```bash
aws lambda create-function \
  --function-name process-orders \
  --runtime python3.9 \
  --handler lambda_function.lambda_handler \
  --role arn:aws:iam::123456789012:role/lambda-execution-role \
  --zip-file fileb:///path/to/function.zip \
  --environment Variables={DB_HOST=database.example.com,API_KEY=secret-key} \
  --region us-east-1
```

---

### delete_function

Delete a Lambda function.

**Required Variables:**
- `function_name` - Function name to delete

**Example:**
```yaml
- name: Delete Lambda function
  module: cloud/aws
  vars:
    command: delete_function
    function_name: old-function
    aws_region: us-east-1
```

**AWS CLI Equivalent:**
```bash
aws lambda delete-function --function-name old-function --region us-east-1
```

---

### invoke_function

Invoke a Lambda function.

**Required Variables:**
- `function_name` - Function name to invoke

**Optional Variables:**
- `payload` - JSON payload (default: {})

**Example:**
```yaml
- name: Invoke Lambda function
  module: cloud/aws
  vars:
    command: invoke_function
    function_name: process-orders
    payload: '{"orderId": "12345", "action": "process"}'
    aws_region: us-east-1
```

**AWS CLI Equivalent:**
```bash
aws lambda invoke \
  --function-name process-orders \
  --payload '{"orderId": "12345", "action": "process"}' \
  --region us-east-1 \
  response.json
```

---

## Common Patterns

### Working with Multiple Regions

```yaml
# Define region list
defaults:
  regions:
    - us-east-1
    - us-west-2
    - eu-west-1

# Loop through regions (when loop support is added)
run:
  - name: List instances in all regions
    module: cloud/aws
    vars:
      command: describe_instances
      aws_region: "{{ item }}"
    loop: "{{ var.regions }}"
```

### Using Facts from Previous Commands

```yaml
run:
  - name: Create VPC
    module: cloud/aws
    vars:
      command: create_vpc
      cidr_block: 10.0.0.0/16
      aws_region: us-east-1

  - name: Create subnet in new VPC
    module: cloud/aws
    vars:
      command: create_subnet
      vpc_id: "{{ facts.vpc_id }}"  # From previous task
      cidr_block: 10.0.1.0/24
      aws_region: us-east-1
```

### Error Handling

```yaml
run:
  - name: Try to create instance
    module: cloud/aws
    vars:
      command: create_instance
      ami_id: ami-12345678
      instance_type: t3.micro
      aws_region: us-east-1
    continue_on_error: true
```
