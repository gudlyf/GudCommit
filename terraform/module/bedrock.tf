resource "aws_bedrockagent_agent" "gudcommit" {
  agent_name                  = "GudCommit"
  agent_resource_role_arn     = aws_iam_role.gudcommit_bedrock.arn
  prepare_agent               = true
  description                 = "Creates clean git commit messages based on output of git diff."
  foundation_model            = var.foundation_model
  idle_session_ttl_in_seconds = 60
  instruction = trimspace(<<-EOT
    Your task is to create clean and comprehensive commit messages as per the conventional commit convention
    by explaining WHAT the changes were and mainly WHY the changes were done. I will provide only an output of
    `git diff --staged` command, and you are to convert it into a commit message. Lines prefixed at the leftmost
    position with a `+` are added lines; lines prefixed with `-` are removed lines. All other lines are only
    context surrounding the code.
    
    Please strictly follow these rules:

    - Do not indicate you are guessing with terms like "likely," "it looks like," or "perhaps."
    - Do not introduce or concluce your response with anything.

    Conventional commit keywords:

    fix, feat, build, chore, ci, docs, style, refactor, perf, test.

    Each line of your response must and only follow the following format, with each line separated by an empty line:

    keyword(full filename being changed): Single sentence describing change and why it was done.
    EOT
  )
  #  prompt_override_configuration {
  #    prompt_configurations = [
  #      {
  #        base_prompt_template = file("${path.module}/prompt_templates/gudcommit_post_processing.json")
  #        inference_configuration = [
  #          {
  #            max_length     = 2048
  #            stop_sequences = ["\n\nHuman:"]
  #            temperature    = 0
  #            top_k          = 250
  #            top_p          = 1
  #          }
  #        ]
  #        parser_mode          = "DEFAULT"
  #        prompt_creation_mode = "OVERRIDDEN"
  #        prompt_state         = "ENABLED"
  #        prompt_type          = "POST_PROCESSING"
  #      }
  #    ]
  #  }
}

resource "aws_bedrockagent_agent_alias" "gudcommit" {
  agent_alias_name = "gudcommit_${data.external.latest_gudcommit_agent_version.result.agent_version != "" ? data.external.latest_gudcommit_agent_version.result.agent_version : "0"}"
  agent_id         = aws_bedrockagent_agent.gudcommit.agent_id
  description      = "GudCommit Version ${data.external.latest_gudcommit_agent_version.result.agent_version != "" ? data.external.latest_gudcommit_agent_version.result.agent_version : "0"}"
}

data "external" "latest_gudcommit_agent_version" {
  program    = ["bash", "${path.module}/scripts/get_agent_details.sh", aws_bedrockagent_agent.gudcommit.agent_id]
  depends_on = [aws_bedrockagent_agent.gudcommit]
}

resource "aws_bedrockagent_agent" "gudchangelog" {
  agent_name                  = "GudChangelog"
  agent_resource_role_arn     = aws_iam_role.gudcommit_bedrock.arn
  prepare_agent               = true
  description                 = "Creates clean CHANGELOG.md content based on merge/pull request diff."
  foundation_model            = var.foundation_model
  idle_session_ttl_in_seconds = 60
  instruction = trimspace(<<-EOT
    Your task is to create a clean and comprehensive `CHANGELOG.md` file content as per the conventional commit convention
    by explaining WHAT the changes were and mainly WHY the changes were done. I will provide only an output of
    `git diff origin/main` command, and you are to assess the changes to create Markdown-formatted content that will be
    appended to the repository's `CHANGELOG.md` file.
    
    Please strictly follow these rules:

    - Do not indicate you are guessing with terms like "likely," "it looks like," or "perhaps."
    - Do not introduce or concluce your response with anything.
    EOT
  )
  #  prompt_override_configuration {
  #    prompt_configurations = [
  #      {
  #        base_prompt_template = file("${path.module}/prompt_templates/gudchangelog_post_processing.json")
  #        inference_configuration = [
  #          {
  #            max_length     = 2048
  #            stop_sequences = ["\n\nHuman:"]
  #            temperature    = 0
  #            top_k          = 250
  #            top_p          = 1
  #          }
  #        ]
  #        parser_mode          = "DEFAULT"
  #        prompt_creation_mode = "OVERRIDDEN"
  #        prompt_state         = "ENABLED"
  #        prompt_type          = "POST_PROCESSING"
  #      }
  #    ]
  #  }
}

resource "aws_bedrockagent_agent_alias" "gudchangelog" {
  agent_alias_name = "gudchangelog_${data.external.latest_gudchangelog_agent_version.result.agent_version != "" ? data.external.latest_gudchangelog_agent_version.result.agent_version : "0"}"
  agent_id         = aws_bedrockagent_agent.gudchangelog.agent_id
  description      = "GudChangelog Version ${data.external.latest_gudchangelog_agent_version.result.agent_version != "" ? data.external.latest_gudchangelog_agent_version.result.agent_version : "0"}"
}

data "external" "latest_gudchangelog_agent_version" {
  program    = ["bash", "${path.module}/scripts/get_agent_details.sh", aws_bedrockagent_agent.gudchangelog.agent_id]
  depends_on = [aws_bedrockagent_agent.gudchangelog]
}

## These are temporarily necessary until AWS supports agent provisioning properly

resource "null_resource" "prepare_gudcommit" {
  triggers = {
    agent_state = sha256(jsonencode(aws_bedrockagent_agent.gudcommit))
  }
  provisioner "local-exec" {
    command = "aws --profile default bedrock-agent prepare-agent --agent-id ${aws_bedrockagent_agent.gudcommit.id}"
  }
  depends_on = [
    aws_bedrockagent_agent.gudcommit,
  ]
}

resource "null_resource" "prepare_gudchangelog" {
  triggers = {
    agent_state = sha256(jsonencode(aws_bedrockagent_agent.gudchangelog))
  }
  provisioner "local-exec" {
    command = "aws --profile default bedrock-agent prepare-agent --agent-id ${aws_bedrockagent_agent.gudchangelog.id}"
  }
  depends_on = [
    aws_bedrockagent_agent.gudchangelog,
  ]
}
