# AWS Cloud Management Module

Comprehensive AWS (Amazon Web Services) management module for OpenFroyo, providing automation for EC2, S3, VPC, IAM, RDS, Lambda, and other AWS services.

## Overview

This module provides agentless AWS cloud infrastructure management using the AWS CLI. It supports 40+ operations across multiple AWS services, allowing you to automate cloud resource provisioning, configuration, and management through OpenFroyo stacks.

## Features

- **EC2 Instance Management**: Create, terminate, start, stop, reboot instances and manage AMIs
- **S3 Storage**: Create buckets, upload/download objects, manage bucket policies
- **VPC and Networking**: Create VPCs, subnets, security groups, and configure ingress rules
- **IAM Management**: Create users, roles, policies, and access keys
- **RDS Databases**: Create, delete, snapshot, and restore database instances
- **Lambda Functions**: Create, delete, and invoke serverless functions
- **Multi-region Support**: Execute operations in any AWS region
- **Flexible Authentication**: Support for credentials, profiles, or IAM roles

## Prerequisites

### Required on Target Hosts

1. **AWS CLI v2** (recommended) or v1
   ```bash
   # Install AWS CLI v2 on Linux
   curl "https://awscli.amazonaws.com/awscli-exe-linux-x86_64.zip" -o "awscliv2.zip"
   unzip awscliv2.zip
   sudo ./aws/install

   # Verify installation
   aws --version
   ```

2. **AWS Credentials** (one of the following):
   - Environment variables (`AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY`)
   - AWS CLI profile configured via `aws configure`
   - IAM instance role (for EC2 instances)

### Setting Up AWS Credentials

#### Option 1: Using AWS CLI Profile (Recommended)
```bash
# Configure default profile
aws configure
# AWS Access Key ID: YOUR_ACCESS_KEY
# AWS Secret Access Key: YOUR_SECRET_KEY
# Default region name: us-east-1
# Default output format: json

# Configure named profile
aws configure --profile production
```

#### Option 2: Environment Variables
```bash
export AWS_ACCESS_KEY_ID="YOUR_ACCESS_KEY"
export AWS_SECRET_ACCESS_KEY="YOUR_SECRET_KEY"
export AWS_DEFAULT_REGION="us-east-1"
```

#### Option 3: IAM Instance Role (For EC2)
When running on EC2 instances, attach an IAM role with appropriate permissions. No credentials needed.

## Quick Start

### Example 1: Launch EC2 Instance

```yaml
# stacks/launch_web_server.ofy
inventory:
  hosts:
    - management-01

run:
  - name: Create EC2 web server
    module: cloud/aws
    vars:
      command: create_instance
      aws_region: us-east-1
      ami_id: ami-0c55b159cbfafe1f0
      instance_type: t3.small
      key_name: my-keypair
      subnet_id: subnet-12345678
      security_group_ids:
        - sg-12345678
```

### Example 2: Create S3 Bucket and Upload Files

```yaml
# stacks/setup_s3_storage.ofy
inventory:
  hosts:
    - management-01

run:
  - name: Create S3 bucket
    module: cloud/aws
    vars:
      command: create_bucket
      bucket_name: my-application-assets
      aws_region: us-west-2

  - name: Upload application assets
    module: cloud/aws
    vars:
      command: upload_object
      bucket_name: my-application-assets
      local_file: /app/assets/index.html
      object_key: public/index.html
      aws_region: us-west-2
```

### Example 3: Create VPC with Subnet

```yaml
# stacks/setup_vpc.ofy
inventory:
  hosts:
    - management-01

run:
  - name: Create VPC
    module: cloud/aws
    vars:
      command: create_vpc
      cidr_block: 10.0.0.0/16
      aws_region: us-east-1

  - name: Create subnet
    module: cloud/aws
    vars:
      command: create_subnet
      vpc_id: "{{ facts.vpc_id }}"
      cidr_block: 10.0.1.0/24
      availability_zone: us-east-1a
      aws_region: us-east-1
```

## Supported Commands

### EC2 Management (10 commands)
- `create_instance` - Launch new EC2 instances
- `terminate_instance` - Terminate EC2 instances
- `start_instance` - Start stopped instances
- `stop_instance` - Stop running instances
- `reboot_instance` - Reboot instances
- `describe_instances` - Get instance details
- `create_ami` - Create AMI from instance
- `describe_amis` - List available AMIs
- `create_tags` - Add tags to instances
- `describe_instance_types` - List available instance types

### S3 Management (8 commands)
- `create_bucket` - Create S3 bucket
- `delete_bucket` - Delete S3 bucket
- `list_buckets` - List all buckets
- `upload_object` - Upload file to S3
- `download_object` - Download file from S3
- `delete_object` - Delete S3 object
- `list_objects` - List objects in bucket
- `set_bucket_policy` - Set bucket policy

### VPC and Networking (8 commands)
- `create_vpc` - Create VPC
- `delete_vpc` - Delete VPC
- `create_subnet` - Create subnet
- `delete_subnet` - Delete subnet
- `create_security_group` - Create security group
- `delete_security_group` - Delete security group
- `authorize_security_group_ingress` - Add ingress rule
- `describe_security_groups` - List security groups

### IAM Management (6 commands)
- `create_user` - Create IAM user
- `delete_user` - Delete IAM user
- `create_role` - Create IAM role
- `delete_role` - Delete IAM role
- `attach_role_policy` - Attach policy to role
- `create_access_key` - Create access key for user

### RDS Management (5 commands)
- `create_db_instance` - Create RDS database
- `delete_db_instance` - Delete RDS database
- `describe_db_instances` - List database instances
- `create_db_snapshot` - Create database snapshot
- `restore_db_instance` - Restore from snapshot

### Lambda Functions (3 commands)
- `create_function` - Create Lambda function
- `delete_function` - Delete Lambda function
- `invoke_function` - Invoke Lambda function

See [COMMANDS.md](COMMANDS.md) for detailed documentation of all commands.

## Module Variables

### Common Variables
- `command` (required) - AWS operation to execute
- `aws_region` (default: us-east-1) - AWS region
- `aws_profile` (default: default) - AWS CLI profile
- `aws_access_key_id` (optional) - AWS access key
- `aws_secret_access_key` (optional) - AWS secret key
- `output_format` (default: json) - AWS CLI output format

### EC2 Variables
- `instance_type` - EC2 instance type (default: t3.micro)
- `ami_id` - AMI ID for instance creation
- `instance_id` - Instance ID for operations
- `key_name` - SSH key pair name
- `subnet_id` - VPC subnet ID
- `security_group_ids` - List of security group IDs

### S3 Variables
- `bucket_name` - S3 bucket name
- `object_key` - S3 object key/path
- `local_file` - Local file path
- `prefix` - Object prefix for listing

### VPC Variables
- `vpc_id` - VPC identifier
- `cidr_block` - CIDR block
- `availability_zone` - Availability zone
- `security_group_name` - Security group name

### IAM Variables
- `user_name` - IAM user name
- `role_name` - IAM role name
- `policy_arn` - Policy ARN

### RDS Variables
- `db_instance_identifier` - Database instance ID
- `db_instance_class` - Database instance class
- `engine` - Database engine (mysql, postgres, etc.)
- `master_username` - Master username
- `master_password` - Master password

### Lambda Variables
- `function_name` - Lambda function name
- `runtime` - Lambda runtime (python3.9, nodejs18.x, etc.)
- `handler` - Function handler
- `zip_file` - Path to deployment package
- `lambda_role` - IAM role ARN

## Security Best Practices

### 1. Use IAM Roles When Possible
For EC2 instances, use IAM instance roles instead of embedding credentials:
```yaml
# No credentials needed - uses instance role
vars:
  command: describe_instances
  aws_region: us-east-1
```

### 2. Use Named Profiles
Store credentials securely with named profiles:
```yaml
vars:
  command: create_instance
  aws_profile: production  # Uses ~/.aws/credentials
  aws_region: us-east-1
```

### 3. Least Privilege Principle
Grant only required permissions. Example IAM policy for OpenFroyo:
```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "ec2:RunInstances",
        "ec2:DescribeInstances",
        "ec2:StopInstances",
        "ec2:StartInstances",
        "s3:CreateBucket",
        "s3:PutObject",
        "s3:GetObject"
      ],
      "Resource": "*"
    }
  ]
}
```

### 4. Protect Sensitive Variables
Use OpenFroyo's variable encryption for sensitive data:
```yaml
vars:
  master_password: "{{ vault.db_password }}"
```

### 5. Enable CloudTrail
Monitor all AWS API calls:
```bash
aws cloudtrail create-trail --name openfroyo-audit --s3-bucket-name audit-logs
```

### 6. Use MFA for Critical Operations
Require MFA for production environments:
```bash
aws configure set mfa_serial arn:aws:iam::123456789012:mfa/user
```

## Cost Optimization

### 1. Use Appropriate Instance Types
- Development: t3.micro, t3.small
- Production: t3.medium, c5.large (based on workload)
- Use Spot instances for non-critical workloads

### 2. Enable Auto-Scaling
Configure auto-scaling groups to match demand:
```yaml
# Scale based on CPU utilization
min_size: 2
max_size: 10
target_cpu: 70
```

### 3. Use Reserved Instances
For predictable workloads, purchase Reserved Instances (up to 75% savings).

### 4. Lifecycle Policies
Delete old snapshots and unused resources:
```yaml
# Delete snapshots older than 30 days
retention_days: 30
```

### 5. Monitor with Cost Explorer
Enable AWS Cost Explorer and set up budgets:
```bash
aws budgets create-budget --account-id 123456789012 --budget file://budget.json
```

### 6. Use S3 Storage Classes
- Frequently accessed: S3 Standard
- Infrequent access: S3 IA
- Archive: S3 Glacier

## Troubleshooting

### AWS CLI Not Found
```bash
# Check if AWS CLI is installed
which aws

# Install if missing
curl "https://awscli.amazonaws.com/awscli-exe-linux-x86_64.zip" -o "awscliv2.zip"
unzip awscliv2.zip
sudo ./aws/install
```

### Authentication Errors
```bash
# Verify credentials
aws sts get-caller-identity

# Check profile
aws configure list --profile production

# Test access
aws ec2 describe-instances --region us-east-1 --profile production
```

### Region-Specific Issues
Some resources are region-specific. Ensure correct region:
```yaml
vars:
  aws_region: us-west-2  # Explicit region
```

### Permission Denied
Check IAM permissions:
```bash
# Get current user
aws sts get-caller-identity

# List attached policies
aws iam list-attached-user-policies --user-name your-user
```

### Rate Limiting
AWS API has rate limits. Implement retry logic or reduce parallelism:
```yaml
strategy: serial  # Execute one host at a time
max_parallel: 5   # Limit parallel operations
```

## Building the Module

```bash
cd modules/cloud/aws/wasm
make build
```

Requirements:
- TinyGo (for WASM compilation)
- Go 1.19+

## Testing

See [test.ofy](test.ofy) for comprehensive test examples.

```bash
# Run stack tests
froyo apply modules/cloud/aws/test.ofy --check
```

## Examples and Workflows

See [WORKFLOWS.md](WORKFLOWS.md) for complete deployment scenarios:
- Setting up a VPC from scratch
- Deploying a web application with EC2 and RDS
- Creating a serverless API with Lambda and API Gateway
- Setting up a static website with S3 and CloudFront

## Additional Resources

- [AWS CLI Documentation](https://docs.aws.amazon.com/cli/)
- [AWS Best Practices](https://aws.amazon.com/architecture/well-architected/)
- [OpenFroyo Documentation](../../docs/)

## License

Part of the OpenFroyo project.
