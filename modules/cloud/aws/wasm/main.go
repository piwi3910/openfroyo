package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
)

// TaskInput represents the input to this module
type TaskInput struct {
	Vars    map[string]interface{} `json:"vars"`
	Context TaskContext            `json:"context"`
}

// TaskContext provides execution context
type TaskContext struct {
	Host     string `json:"host"`
	TaskName string `json:"task_name"`
}

// TaskOutput represents the output from this module
type TaskOutput struct {
	Status  string                 `json:"status"`
	Message string                 `json:"message"`
	Facts   map[string]interface{} `json:"facts"`
}

func main() {
	// Read input from stdin
	inputBytes, err := io.ReadAll(os.Stdin)
	if err != nil {
		outputError(fmt.Sprintf("Failed to read stdin: %v", err))
		return
	}

	var input TaskInput
	if err := json.Unmarshal(inputBytes, &input); err != nil {
		outputError(fmt.Sprintf("Failed to parse input JSON: %v", err))
		return
	}

	// Extract common parameters
	command := getString(input.Vars, "command", "")
	if command == "" {
		outputError("Missing required variable: command")
		return
	}

	awsRegion := getString(input.Vars, "aws_region", "us-east-1")
	awsProfile := getString(input.Vars, "aws_profile", "default")
	outputFormat := getString(input.Vars, "output_format", "json")

	// Build AWS CLI command based on operation
	result := buildAWSCommand(command, input.Vars, awsRegion, awsProfile, outputFormat)
	printOutput(result)
}

func buildAWSCommand(command string, vars map[string]interface{}, region, profile, outputFormat string) TaskOutput {
	var shellExec []map[string]interface{}
	var cmdStr string
	var err error

	// Route to appropriate command builder
	switch {
	// EC2 Commands
	case command == "create_instance":
		cmdStr, err = buildEC2CreateInstance(vars, region)
	case command == "terminate_instance":
		cmdStr, err = buildEC2TerminateInstance(vars, region)
	case command == "start_instance":
		cmdStr, err = buildEC2StartInstance(vars, region)
	case command == "stop_instance":
		cmdStr, err = buildEC2StopInstance(vars, region)
	case command == "reboot_instance":
		cmdStr, err = buildEC2RebootInstance(vars, region)
	case command == "describe_instances":
		cmdStr, err = buildEC2DescribeInstances(vars, region)
	case command == "create_ami":
		cmdStr, err = buildEC2CreateAMI(vars, region)
	case command == "describe_amis":
		cmdStr, err = buildEC2DescribeAMIs(vars, region)
	case command == "create_tags":
		cmdStr, err = buildEC2CreateTags(vars, region)
	case command == "describe_instance_types":
		cmdStr, err = buildEC2DescribeInstanceTypes(vars, region)

	// S3 Commands
	case command == "create_bucket":
		cmdStr, err = buildS3CreateBucket(vars, region)
	case command == "delete_bucket":
		cmdStr, err = buildS3DeleteBucket(vars, region)
	case command == "list_buckets":
		cmdStr, err = buildS3ListBuckets(vars, region)
	case command == "upload_object":
		cmdStr, err = buildS3UploadObject(vars, region)
	case command == "download_object":
		cmdStr, err = buildS3DownloadObject(vars, region)
	case command == "delete_object":
		cmdStr, err = buildS3DeleteObject(vars, region)
	case command == "list_objects":
		cmdStr, err = buildS3ListObjects(vars, region)
	case command == "set_bucket_policy":
		cmdStr, err = buildS3SetBucketPolicy(vars, region)

	// VPC and Networking Commands
	case command == "create_vpc":
		cmdStr, err = buildVPCCreate(vars, region)
	case command == "delete_vpc":
		cmdStr, err = buildVPCDelete(vars, region)
	case command == "create_subnet":
		cmdStr, err = buildSubnetCreate(vars, region)
	case command == "delete_subnet":
		cmdStr, err = buildSubnetDelete(vars, region)
	case command == "create_security_group":
		cmdStr, err = buildSecurityGroupCreate(vars, region)
	case command == "delete_security_group":
		cmdStr, err = buildSecurityGroupDelete(vars, region)
	case command == "authorize_security_group_ingress":
		cmdStr, err = buildSecurityGroupAuthorizeIngress(vars, region)
	case command == "describe_security_groups":
		cmdStr, err = buildSecurityGroupDescribe(vars, region)

	// IAM Commands
	case command == "create_user":
		cmdStr, err = buildIAMCreateUser(vars, region)
	case command == "delete_user":
		cmdStr, err = buildIAMDeleteUser(vars, region)
	case command == "create_role":
		cmdStr, err = buildIAMCreateRole(vars, region)
	case command == "delete_role":
		cmdStr, err = buildIAMDeleteRole(vars, region)
	case command == "attach_role_policy":
		cmdStr, err = buildIAMAttachRolePolicy(vars, region)
	case command == "create_access_key":
		cmdStr, err = buildIAMCreateAccessKey(vars, region)

	// RDS Commands
	case command == "create_db_instance":
		cmdStr, err = buildRDSCreateInstance(vars, region)
	case command == "delete_db_instance":
		cmdStr, err = buildRDSDeleteInstance(vars, region)
	case command == "describe_db_instances":
		cmdStr, err = buildRDSDescribeInstances(vars, region)
	case command == "create_db_snapshot":
		cmdStr, err = buildRDSCreateSnapshot(vars, region)
	case command == "restore_db_instance":
		cmdStr, err = buildRDSRestoreInstance(vars, region)

	// Lambda Commands
	case command == "create_function":
		cmdStr, err = buildLambdaCreateFunction(vars, region)
	case command == "delete_function":
		cmdStr, err = buildLambdaDeleteFunction(vars, region)
	case command == "invoke_function":
		cmdStr, err = buildLambdaInvokeFunction(vars, region)

	default:
		return TaskOutput{
			Status:  "failed",
			Message: fmt.Sprintf("Unknown command: %s", command),
			Facts:   make(map[string]interface{}),
		}
	}

	if err != nil {
		return TaskOutput{
			Status:  "failed",
			Message: err.Error(),
			Facts:   make(map[string]interface{}),
		}
	}

	// Add AWS credentials setup if provided
	awsAccessKey := getString(vars, "aws_access_key_id", "")
	awsSecretKey := getString(vars, "aws_secret_access_key", "")

	if awsAccessKey != "" && awsSecretKey != "" {
		shellExec = append(shellExec, map[string]interface{}{
			"type":    "shell",
			"command": fmt.Sprintf("export AWS_ACCESS_KEY_ID='%s'", awsAccessKey),
		})
		shellExec = append(shellExec, map[string]interface{}{
			"type":    "shell",
			"command": fmt.Sprintf("export AWS_SECRET_ACCESS_KEY='%s'", awsSecretKey),
		})
	} else if profile != "default" {
		cmdStr = fmt.Sprintf("%s --profile %s", cmdStr, profile)
	}

	// Add the main AWS CLI command
	shellExec = append(shellExec, map[string]interface{}{
		"type":    "shell",
		"command": cmdStr,
	})

	return TaskOutput{
		Status:  "ok",
		Message: fmt.Sprintf("Executing AWS %s command", command),
		Facts: map[string]interface{}{
			"shell_exec": shellExec,
			"command":    command,
			"region":     region,
		},
	}
}

// ==================== EC2 Command Builders ====================

func buildEC2CreateInstance(vars map[string]interface{}, region string) (string, error) {
	amiID := getString(vars, "ami_id", "")
	if amiID == "" {
		return "", fmt.Errorf("ami_id is required for create_instance")
	}

	instanceType := getString(vars, "instance_type", "t3.micro")
	keyName := getString(vars, "key_name", "")
	subnetID := getString(vars, "subnet_id", "")

	cmd := fmt.Sprintf("aws ec2 run-instances --image-id %s --instance-type %s --region %s --output json",
		amiID, instanceType, region)

	if keyName != "" {
		cmd += fmt.Sprintf(" --key-name %s", keyName)
	}

	if subnetID != "" {
		cmd += fmt.Sprintf(" --subnet-id %s", subnetID)
	}

	// Handle security groups
	if sgIDs, ok := vars["security_group_ids"].([]interface{}); ok && len(sgIDs) > 0 {
		var sgList []string
		for _, sg := range sgIDs {
			if sgStr, ok := sg.(string); ok {
				sgList = append(sgList, sgStr)
			}
		}
		if len(sgList) > 0 {
			cmd += fmt.Sprintf(" --security-group-ids %s", strings.Join(sgList, " "))
		}
	}

	return cmd, nil
}

func buildEC2TerminateInstance(vars map[string]interface{}, region string) (string, error) {
	instanceID := getString(vars, "instance_id", "")
	if instanceID == "" {
		return "", fmt.Errorf("instance_id is required for terminate_instance")
	}

	return fmt.Sprintf("aws ec2 terminate-instances --instance-ids %s --region %s --output json",
		instanceID, region), nil
}

func buildEC2StartInstance(vars map[string]interface{}, region string) (string, error) {
	instanceID := getString(vars, "instance_id", "")
	if instanceID == "" {
		return "", fmt.Errorf("instance_id is required for start_instance")
	}

	return fmt.Sprintf("aws ec2 start-instances --instance-ids %s --region %s --output json",
		instanceID, region), nil
}

func buildEC2StopInstance(vars map[string]interface{}, region string) (string, error) {
	instanceID := getString(vars, "instance_id", "")
	if instanceID == "" {
		return "", fmt.Errorf("instance_id is required for stop_instance")
	}

	return fmt.Sprintf("aws ec2 stop-instances --instance-ids %s --region %s --output json",
		instanceID, region), nil
}

func buildEC2RebootInstance(vars map[string]interface{}, region string) (string, error) {
	instanceID := getString(vars, "instance_id", "")
	if instanceID == "" {
		return "", fmt.Errorf("instance_id is required for reboot_instance")
	}

	return fmt.Sprintf("aws ec2 reboot-instances --instance-ids %s --region %s",
		instanceID, region), nil
}

func buildEC2DescribeInstances(vars map[string]interface{}, region string) (string, error) {
	cmd := fmt.Sprintf("aws ec2 describe-instances --region %s --output json", region)

	instanceID := getString(vars, "instance_id", "")
	if instanceID != "" {
		cmd += fmt.Sprintf(" --instance-ids %s", instanceID)
	}

	return cmd, nil
}

func buildEC2CreateAMI(vars map[string]interface{}, region string) (string, error) {
	instanceID := getString(vars, "instance_id", "")
	imageName := getString(vars, "image_name", "")

	if instanceID == "" {
		return "", fmt.Errorf("instance_id is required for create_ami")
	}
	if imageName == "" {
		return "", fmt.Errorf("image_name is required for create_ami")
	}

	return fmt.Sprintf("aws ec2 create-image --instance-id %s --name '%s' --region %s --output json",
		instanceID, imageName, region), nil
}

func buildEC2DescribeAMIs(vars map[string]interface{}, region string) (string, error) {
	return fmt.Sprintf("aws ec2 describe-images --owners self --region %s --output json", region), nil
}

func buildEC2CreateTags(vars map[string]interface{}, region string) (string, error) {
	instanceID := getString(vars, "instance_id", "")
	if instanceID == "" {
		return "", fmt.Errorf("instance_id is required for create_tags")
	}

	tags, ok := vars["tag_specifications"].(map[string]interface{})
	if !ok || len(tags) == 0 {
		return "", fmt.Errorf("tag_specifications is required for create_tags")
	}

	var tagPairs []string
	for key, value := range tags {
		tagPairs = append(tagPairs, fmt.Sprintf("Key=%s,Value=%v", key, value))
	}

	return fmt.Sprintf("aws ec2 create-tags --resources %s --tags %s --region %s",
		instanceID, strings.Join(tagPairs, " "), region), nil
}

func buildEC2DescribeInstanceTypes(vars map[string]interface{}, region string) (string, error) {
	return fmt.Sprintf("aws ec2 describe-instance-types --region %s --output json", region), nil
}

// ==================== S3 Command Builders ====================

func buildS3CreateBucket(vars map[string]interface{}, region string) (string, error) {
	bucketName := getString(vars, "bucket_name", "")
	if bucketName == "" {
		return "", fmt.Errorf("bucket_name is required for create_bucket")
	}

	if region == "us-east-1" {
		return fmt.Sprintf("aws s3 mb s3://%s", bucketName), nil
	}

	return fmt.Sprintf("aws s3 mb s3://%s --region %s", bucketName, region), nil
}

func buildS3DeleteBucket(vars map[string]interface{}, region string) (string, error) {
	bucketName := getString(vars, "bucket_name", "")
	if bucketName == "" {
		return "", fmt.Errorf("bucket_name is required for delete_bucket")
	}

	return fmt.Sprintf("aws s3 rb s3://%s --force", bucketName), nil
}

func buildS3ListBuckets(vars map[string]interface{}, region string) (string, error) {
	return "aws s3 ls", nil
}

func buildS3UploadObject(vars map[string]interface{}, region string) (string, error) {
	bucketName := getString(vars, "bucket_name", "")
	localFile := getString(vars, "local_file", "")
	objectKey := getString(vars, "object_key", "")

	if bucketName == "" {
		return "", fmt.Errorf("bucket_name is required for upload_object")
	}
	if localFile == "" {
		return "", fmt.Errorf("local_file is required for upload_object")
	}

	if objectKey == "" {
		objectKey = localFile
	}

	return fmt.Sprintf("aws s3 cp '%s' s3://%s/%s --region %s",
		localFile, bucketName, objectKey, region), nil
}

func buildS3DownloadObject(vars map[string]interface{}, region string) (string, error) {
	bucketName := getString(vars, "bucket_name", "")
	localFile := getString(vars, "local_file", "")
	objectKey := getString(vars, "object_key", "")

	if bucketName == "" {
		return "", fmt.Errorf("bucket_name is required for download_object")
	}
	if objectKey == "" {
		return "", fmt.Errorf("object_key is required for download_object")
	}
	if localFile == "" {
		return "", fmt.Errorf("local_file is required for download_object")
	}

	return fmt.Sprintf("aws s3 cp s3://%s/%s '%s' --region %s",
		bucketName, objectKey, localFile, region), nil
}

func buildS3DeleteObject(vars map[string]interface{}, region string) (string, error) {
	bucketName := getString(vars, "bucket_name", "")
	objectKey := getString(vars, "object_key", "")

	if bucketName == "" {
		return "", fmt.Errorf("bucket_name is required for delete_object")
	}
	if objectKey == "" {
		return "", fmt.Errorf("object_key is required for delete_object")
	}

	return fmt.Sprintf("aws s3 rm s3://%s/%s --region %s",
		bucketName, objectKey, region), nil
}

func buildS3ListObjects(vars map[string]interface{}, region string) (string, error) {
	bucketName := getString(vars, "bucket_name", "")
	if bucketName == "" {
		return "", fmt.Errorf("bucket_name is required for list_objects")
	}

	prefix := getString(vars, "prefix", "")
	cmd := fmt.Sprintf("aws s3 ls s3://%s", bucketName)

	if prefix != "" {
		cmd += fmt.Sprintf("/%s", prefix)
	}

	cmd += " --recursive"

	return cmd, nil
}

func buildS3SetBucketPolicy(vars map[string]interface{}, region string) (string, error) {
	bucketName := getString(vars, "bucket_name", "")
	bucketPolicy := getString(vars, "bucket_policy", "")

	if bucketName == "" {
		return "", fmt.Errorf("bucket_name is required for set_bucket_policy")
	}
	if bucketPolicy == "" {
		return "", fmt.Errorf("bucket_policy is required for set_bucket_policy")
	}

	return fmt.Sprintf("aws s3api put-bucket-policy --bucket %s --policy '%s' --region %s",
		bucketName, bucketPolicy, region), nil
}

// ==================== VPC Command Builders ====================

func buildVPCCreate(vars map[string]interface{}, region string) (string, error) {
	cidrBlock := getString(vars, "cidr_block", "")
	if cidrBlock == "" {
		return "", fmt.Errorf("cidr_block is required for create_vpc")
	}

	return fmt.Sprintf("aws ec2 create-vpc --cidr-block %s --region %s --output json",
		cidrBlock, region), nil
}

func buildVPCDelete(vars map[string]interface{}, region string) (string, error) {
	vpcID := getString(vars, "vpc_id", "")
	if vpcID == "" {
		return "", fmt.Errorf("vpc_id is required for delete_vpc")
	}

	return fmt.Sprintf("aws ec2 delete-vpc --vpc-id %s --region %s",
		vpcID, region), nil
}

func buildSubnetCreate(vars map[string]interface{}, region string) (string, error) {
	vpcID := getString(vars, "vpc_id", "")
	cidrBlock := getString(vars, "cidr_block", "")

	if vpcID == "" {
		return "", fmt.Errorf("vpc_id is required for create_subnet")
	}
	if cidrBlock == "" {
		return "", fmt.Errorf("cidr_block is required for create_subnet")
	}

	cmd := fmt.Sprintf("aws ec2 create-subnet --vpc-id %s --cidr-block %s --region %s --output json",
		vpcID, cidrBlock, region)

	az := getString(vars, "availability_zone", "")
	if az != "" {
		cmd += fmt.Sprintf(" --availability-zone %s", az)
	}

	return cmd, nil
}

func buildSubnetDelete(vars map[string]interface{}, region string) (string, error) {
	subnetID := getString(vars, "subnet_id", "")
	if subnetID == "" {
		return "", fmt.Errorf("subnet_id is required for delete_subnet")
	}

	return fmt.Sprintf("aws ec2 delete-subnet --subnet-id %s --region %s",
		subnetID, region), nil
}

func buildSecurityGroupCreate(vars map[string]interface{}, region string) (string, error) {
	groupName := getString(vars, "security_group_name", "")
	description := getString(vars, "security_group_description", "")
	vpcID := getString(vars, "vpc_id", "")

	if groupName == "" {
		return "", fmt.Errorf("security_group_name is required for create_security_group")
	}
	if description == "" {
		description = groupName
	}

	cmd := fmt.Sprintf("aws ec2 create-security-group --group-name '%s' --description '%s' --region %s --output json",
		groupName, description, region)

	if vpcID != "" {
		cmd += fmt.Sprintf(" --vpc-id %s", vpcID)
	}

	return cmd, nil
}

func buildSecurityGroupDelete(vars map[string]interface{}, region string) (string, error) {
	sgID := getString(vars, "security_group_id", "")
	if sgID == "" {
		return "", fmt.Errorf("security_group_id is required for delete_security_group")
	}

	return fmt.Sprintf("aws ec2 delete-security-group --group-id %s --region %s",
		sgID, region), nil
}

func buildSecurityGroupAuthorizeIngress(vars map[string]interface{}, region string) (string, error) {
	sgID := getString(vars, "security_group_id", "")
	if sgID == "" {
		return "", fmt.Errorf("security_group_id is required for authorize_security_group_ingress")
	}

	protocol := getString(vars, "ip_protocol", "tcp")
	fromPort := getInt(vars, "from_port", 0)
	toPort := getInt(vars, "to_port", 0)
	cidrIP := getString(vars, "cidr_ip", "0.0.0.0/0")

	if fromPort == 0 || toPort == 0 {
		return "", fmt.Errorf("from_port and to_port are required for authorize_security_group_ingress")
	}

	return fmt.Sprintf("aws ec2 authorize-security-group-ingress --group-id %s --protocol %s --port %d-%d --cidr %s --region %s",
		sgID, protocol, fromPort, toPort, cidrIP, region), nil
}

func buildSecurityGroupDescribe(vars map[string]interface{}, region string) (string, error) {
	cmd := fmt.Sprintf("aws ec2 describe-security-groups --region %s --output json", region)

	sgID := getString(vars, "security_group_id", "")
	if sgID != "" {
		cmd += fmt.Sprintf(" --group-ids %s", sgID)
	}

	return cmd, nil
}

// ==================== IAM Command Builders ====================

func buildIAMCreateUser(vars map[string]interface{}, region string) (string, error) {
	userName := getString(vars, "user_name", "")
	if userName == "" {
		return "", fmt.Errorf("user_name is required for create_user")
	}

	return fmt.Sprintf("aws iam create-user --user-name %s --output json", userName), nil
}

func buildIAMDeleteUser(vars map[string]interface{}, region string) (string, error) {
	userName := getString(vars, "user_name", "")
	if userName == "" {
		return "", fmt.Errorf("user_name is required for delete_user")
	}

	return fmt.Sprintf("aws iam delete-user --user-name %s", userName), nil
}

func buildIAMCreateRole(vars map[string]interface{}, region string) (string, error) {
	roleName := getString(vars, "role_name", "")
	assumeRolePolicy := getString(vars, "assume_role_policy_document", "")

	if roleName == "" {
		return "", fmt.Errorf("role_name is required for create_role")
	}
	if assumeRolePolicy == "" {
		return "", fmt.Errorf("assume_role_policy_document is required for create_role")
	}

	// Check if policy is a file path or inline JSON
	var policyParam string
	if strings.HasPrefix(assumeRolePolicy, "file://") {
		policyParam = fmt.Sprintf("--assume-role-policy-document %s", assumeRolePolicy)
	} else {
		policyParam = fmt.Sprintf("--assume-role-policy-document '%s'", assumeRolePolicy)
	}

	return fmt.Sprintf("aws iam create-role --role-name %s %s --output json", roleName, policyParam), nil
}

func buildIAMDeleteRole(vars map[string]interface{}, region string) (string, error) {
	roleName := getString(vars, "role_name", "")
	if roleName == "" {
		return "", fmt.Errorf("role_name is required for delete_role")
	}

	return fmt.Sprintf("aws iam delete-role --role-name %s", roleName), nil
}

func buildIAMAttachRolePolicy(vars map[string]interface{}, region string) (string, error) {
	roleName := getString(vars, "role_name", "")
	policyARN := getString(vars, "policy_arn", "")

	if roleName == "" {
		return "", fmt.Errorf("role_name is required for attach_role_policy")
	}
	if policyARN == "" {
		return "", fmt.Errorf("policy_arn is required for attach_role_policy")
	}

	return fmt.Sprintf("aws iam attach-role-policy --role-name %s --policy-arn %s", roleName, policyARN), nil
}

func buildIAMCreateAccessKey(vars map[string]interface{}, region string) (string, error) {
	userName := getString(vars, "user_name", "")
	if userName == "" {
		return "", fmt.Errorf("user_name is required for create_access_key")
	}

	return fmt.Sprintf("aws iam create-access-key --user-name %s --output json", userName), nil
}

// ==================== RDS Command Builders ====================

func buildRDSCreateInstance(vars map[string]interface{}, region string) (string, error) {
	dbIdentifier := getString(vars, "db_instance_identifier", "")
	dbClass := getString(vars, "db_instance_class", "db.t3.micro")
	engine := getString(vars, "engine", "mysql")
	masterUser := getString(vars, "master_username", "")
	masterPass := getString(vars, "master_password", "")
	allocatedStorage := getInt(vars, "allocated_storage", 20)

	if dbIdentifier == "" {
		return "", fmt.Errorf("db_instance_identifier is required for create_db_instance")
	}
	if masterUser == "" {
		return "", fmt.Errorf("master_username is required for create_db_instance")
	}
	if masterPass == "" {
		return "", fmt.Errorf("master_password is required for create_db_instance")
	}

	return fmt.Sprintf("aws rds create-db-instance --db-instance-identifier %s --db-instance-class %s --engine %s --master-username %s --master-user-password '%s' --allocated-storage %d --region %s --output json",
		dbIdentifier, dbClass, engine, masterUser, masterPass, allocatedStorage, region), nil
}

func buildRDSDeleteInstance(vars map[string]interface{}, region string) (string, error) {
	dbIdentifier := getString(vars, "db_instance_identifier", "")
	if dbIdentifier == "" {
		return "", fmt.Errorf("db_instance_identifier is required for delete_db_instance")
	}

	return fmt.Sprintf("aws rds delete-db-instance --db-instance-identifier %s --skip-final-snapshot --region %s --output json",
		dbIdentifier, region), nil
}

func buildRDSDescribeInstances(vars map[string]interface{}, region string) (string, error) {
	cmd := fmt.Sprintf("aws rds describe-db-instances --region %s --output json", region)

	dbIdentifier := getString(vars, "db_instance_identifier", "")
	if dbIdentifier != "" {
		cmd += fmt.Sprintf(" --db-instance-identifier %s", dbIdentifier)
	}

	return cmd, nil
}

func buildRDSCreateSnapshot(vars map[string]interface{}, region string) (string, error) {
	dbIdentifier := getString(vars, "db_instance_identifier", "")
	snapshotID := getString(vars, "snapshot_identifier", "")

	if dbIdentifier == "" {
		return "", fmt.Errorf("db_instance_identifier is required for create_db_snapshot")
	}
	if snapshotID == "" {
		return "", fmt.Errorf("snapshot_identifier is required for create_db_snapshot")
	}

	return fmt.Sprintf("aws rds create-db-snapshot --db-instance-identifier %s --db-snapshot-identifier %s --region %s --output json",
		dbIdentifier, snapshotID, region), nil
}

func buildRDSRestoreInstance(vars map[string]interface{}, region string) (string, error) {
	dbIdentifier := getString(vars, "db_instance_identifier", "")
	snapshotID := getString(vars, "snapshot_identifier", "")

	if dbIdentifier == "" {
		return "", fmt.Errorf("db_instance_identifier is required for restore_db_instance")
	}
	if snapshotID == "" {
		return "", fmt.Errorf("snapshot_identifier is required for restore_db_instance")
	}

	return fmt.Sprintf("aws rds restore-db-instance-from-db-snapshot --db-instance-identifier %s --db-snapshot-identifier %s --region %s --output json",
		dbIdentifier, snapshotID, region), nil
}

// ==================== Lambda Command Builders ====================

func buildLambdaCreateFunction(vars map[string]interface{}, region string) (string, error) {
	functionName := getString(vars, "function_name", "")
	runtime := getString(vars, "runtime", "python3.9")
	handler := getString(vars, "handler", "index.handler")
	role := getString(vars, "lambda_role", "")
	zipFile := getString(vars, "zip_file", "")

	if functionName == "" {
		return "", fmt.Errorf("function_name is required for create_function")
	}
	if role == "" {
		return "", fmt.Errorf("lambda_role is required for create_function")
	}
	if zipFile == "" {
		return "", fmt.Errorf("zip_file is required for create_function")
	}

	cmd := fmt.Sprintf("aws lambda create-function --function-name %s --runtime %s --handler %s --role %s --zip-file fileb://%s --region %s --output json",
		functionName, runtime, handler, role, zipFile, region)

	// Add environment variables if provided
	if envVars, ok := vars["environment_variables"].(map[string]interface{}); ok && len(envVars) > 0 {
		var envPairs []string
		for key, value := range envVars {
			envPairs = append(envPairs, fmt.Sprintf("%s=%v", key, value))
		}
		cmd += fmt.Sprintf(" --environment Variables={%s}", strings.Join(envPairs, ","))
	}

	return cmd, nil
}

func buildLambdaDeleteFunction(vars map[string]interface{}, region string) (string, error) {
	functionName := getString(vars, "function_name", "")
	if functionName == "" {
		return "", fmt.Errorf("function_name is required for delete_function")
	}

	return fmt.Sprintf("aws lambda delete-function --function-name %s --region %s",
		functionName, region), nil
}

func buildLambdaInvokeFunction(vars map[string]interface{}, region string) (string, error) {
	functionName := getString(vars, "function_name", "")
	if functionName == "" {
		return "", fmt.Errorf("function_name is required for invoke_function")
	}

	payload := getString(vars, "payload", "{}")

	return fmt.Sprintf("aws lambda invoke --function-name %s --payload '%s' --region %s response.json && cat response.json",
		functionName, payload, region), nil
}

// ==================== Helper Functions ====================

func getString(vars map[string]interface{}, key, defaultVal string) string {
	if val, ok := vars[key].(string); ok {
		return val
	}
	return defaultVal
}

func getInt(vars map[string]interface{}, key string, defaultVal int) int {
	if val, ok := vars[key].(float64); ok {
		return int(val)
	}
	if val, ok := vars[key].(int); ok {
		return val
	}
	return defaultVal
}

func outputError(msg string) {
	result := TaskOutput{
		Status:  "failed",
		Message: msg,
		Facts:   make(map[string]interface{}),
	}
	printOutput(result)
}

func printOutput(output TaskOutput) {
	outputJSON, _ := json.Marshal(output)
	fmt.Println(string(outputJSON))
}
