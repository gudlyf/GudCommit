terraform {
  required_version = ">= 1.9"

  required_providers {
    aws = {
      version = "~> 5.64.0"
      source  = "hashicorp/aws"
    }
    external = {
      version = "~> 2.3.3"
      source  = "hashicorp/external"
    }
    null = {
      version = "~> 3.2.2"
      source  = "hashicorp/null"
    }
  }

  backend "s3" {
    bucket         = "terraform-state"
    key            = "gudcommit/terraform.tfstate"
    region         = "us-east-1"
    profile        = "default"
    dynamodb_table = "terraform-lock"
  }
}

provider "aws" {
  region  = "us-east-1"
  profile = "default"

  default_tags {
    tags = {
      ENV = "dev"
      APP = "gudcommit"
    }
  }
}

module "gudcommit" {
  source = "../module"
}
