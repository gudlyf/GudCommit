resource "aws_iam_role" "gudcommit_bedrock" {
  assume_role_policy = data.aws_iam_policy_document.gudcommit_bedrock.json
  name_prefix        = "gudcommit_"
}

data "aws_iam_policy_document" "gudcommit_bedrock" {
  statement {
    actions = ["sts:AssumeRole"]
    principals {
      identifiers = ["bedrock.amazonaws.com"]
      type        = "Service"
    }
    condition {
      test     = "StringEquals"
      values   = [data.aws_caller_identity.current.account_id]
      variable = "aws:SourceAccount"
    }

    condition {
      test     = "ArnLike"
      values   = ["arn:aws:bedrock:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:agent/*"]
      variable = "AWS:SourceArn"
    }
  }
}

data "aws_iam_policy_document" "gudcommit_invoke" {
  statement {
    actions = ["bedrock:InvokeModel"]
    resources = [
      "arn:aws:bedrock:${data.aws_region.current.name}::foundation-model/anthropic.claude-*"
    ]
  }
}

resource "aws_iam_role_policy" "gudcommit_bedrock" {
  policy = data.aws_iam_policy_document.gudcommit_invoke.json
  role   = aws_iam_role.gudcommit_bedrock.id
}