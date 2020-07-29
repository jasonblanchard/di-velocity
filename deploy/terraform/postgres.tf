data "terraform_remote_state" "vpc" {
  backend = "s3"

  config = {
    bucket = "di-terraform"
    key    = "vpc/terraform.tfstate"
    region = "us-east-1"
  }
}

resource "aws_db_subnet_group" "default" {
  name       = "di-velocity"
  subnet_ids = [data.terraform_remote_state.vpc.outputs.subnet_us_east_1a_id, data.terraform_remote_state.vpc.outputs.subnet_us_east_1b_id]
}

variable "db_password" {
  type    = string
}

# Security Group
resource "aws_security_group" "postgres" {
  name        = "DiVelocityPostgres"
  vpc_id      = data.terraform_remote_state.vpc.outputs.vpc_id

  ingress {
    from_port   = 5432
    to_port     = 5432
    protocol    = "tcp"
    security_groups = [data.terraform_remote_state.vpc.outputs.security_group_web_id]
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }
}

resource "aws_rds_cluster" "default" {
  cluster_identifier      = "di-velocity"
  engine                  = "aurora-postgresql"
  db_subnet_group_name    = aws_db_subnet_group.default.name
  engine_version          = "10.7"
  engine_mode             = "serverless"
  database_name           = "di_velocity"
  master_username         = "postgres"
  master_password         = var.db_password
  backup_retention_period = 5
  vpc_security_group_ids  = [aws_security_group.postgres.id]
  storage_encrypted       = true
  skip_final_snapshot     = true

  scaling_configuration {
    auto_pause               = true
    max_capacity             = 2
    min_capacity             = 2
    seconds_until_auto_pause = 300
    timeout_action           = "ForceApplyCapacityChange"
  }
}

output "database_host" {
  value = aws_rds_cluster.default.endpoint
}
