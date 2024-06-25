terraform {
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
