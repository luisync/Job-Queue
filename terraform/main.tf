terraform {
    required_version = ">= 1.0.0"
    required_providers {
        aws = {
            source = "hashicorp/aws"
            version = "~> 5.0"
        }
    }
}

provider "aws" {
    region = "eu-north-1"
}

resource "aws_vpc" "main" {
    cidr_block = "10.0.0.0/16"
    enable_dns_hostnames = true
    enable_dns_support = true

    tags = {
        Name = "jobqueue-vpc"
    }
}


resource "aws_subnet" "public" {
    vpc_id = aws_vpc.main.id
    cidr_block = "10.0.1.0/24"
    map_public_ip_on_launch = true
    availability_zone = "eu-north-1a"

    tags = {
        Name = "jobqueue-public-subnet"
    }
}

resource "aws_internet_gateway" "gw" {
    vpc_id = aws_vpc.main.id

    tags = {
        Name = "jobqueue-igw"
    }
}

resource "aws_route_table" "public" {
    vpc_id = aws_vpc.main.id

    route {
        cidr_block = "0.0.0.0/0"
        gateway_id = aws_internet_gateway.gw.id
    }

    tags = {
        Name = "jobqueue-public-rt"
    }
}

# Allow incoming connections to be corrently pointed to the subnet.
resource "aws_route_table_association" "public" {
    subnet_id = aws_subnet.public.id
    route_table_id = aws_route_table.public.id
}

resource "aws_security_group" "k8s_sg" {
    name = "jobqueue-k8s-sg"
    description = "Security group for K3s cluster"
    vpc_id = aws_vpc.main.id

    # Clusters have inbound messages for APIs and none outbound.
    ingress {
        description = "Kubernetes API"
        from_port = 6443
        to_port = 6443
        protocol = "tcp"
        cidr_blocks = ["0.0.0.0/0"]
    }

    ingress {
        description = "Server API"
        from_port = 80
        to_port = 80
        protocol = "tcp"
        cidr_blocks = ["0.0.0.0/0"]
    }

    egress {
        from_port = 0
        to_port = 0
        protocol = "-1"
        cidr_blocks = ["0.0.0.0/0"]
    }
}

resource "aws_iam_role" "k3s_ssm" { 
    name = "k3s-ssm-role"

    assume_role_policy = jsonencode({
        Version = "2012-10-17"
        Statement = [{
            Effect = "Allow"
            Principal = {
                Service = "ec2.amazonwas.com"
            }
            Action = "sts:AssumeRole"
        }]
    })
}

resource "aws_iam_role_policy_attachment" "k3s_ssm" {
    role = aws_iamrole.k3s_ssm.name
    policy_arn = "arn:aws:iam::aws:policy/AmazonSSMManagedInstanceCore"
}

resource "aws_iam_instance_profile" "k3s_ssm" {
    name = "k3s-ssm-instance-profile"
    role = aws_iam_role.k3s_ssm.name
}

resource "aws_instance" "k8s_node" {
    ami = "ami-0aba19e56f3eaec05"
    instance_type = "t3.micro"
    subnet_id = aws_subnet.public.id
    vpc_security_group_ids = [aws_security_group.k8s_sg.id]
    iam_instance_profile = aws_iam_instance_profile.k3s_ssm.name

    user_data = <<-EOF
                #!/bin/bash
                curl -sfL https://get.k3s.io | sh -s - --tls-san $(curl -s http://checkip.amazonaws.com)
                chmod 644 /etc/rancher/k3s/k3s.yaml
                EOF
    
    tags = {
        Name = "k3s-cluster-node"
    }
}

output "k8s_public_ip" {
    description = "Public IP of the Kubernetes Server"
    value = aws_instance.k8s_node.public_ip
}