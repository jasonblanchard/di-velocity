provider "aws" {
  region = "us-east-1"
}

terraform {
  backend "s3" {
    bucket = "di-terraform"
    key    = "di-velocity/terraform.tfstate"
    region = "us-east-1"
  }
}
