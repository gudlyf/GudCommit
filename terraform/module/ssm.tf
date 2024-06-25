resource "aws_ssm_parameter" "bedrock_gudcommit_agent_id" {
  name  = "/gudcommit/gudcommit_bedrock_agent_id"
  type  = "String"
  value = aws_bedrockagent_agent.gudcommit.agent_id
}

resource "aws_ssm_parameter" "bedrock_gudcommit_agent_alias_id" {
  name  = "/gudcommit/gudcommit_bedrock_agent_alias_id"
  type  = "String"
  value = aws_bedrockagent_agent_alias.gudcommit.agent_alias_id
}

resource "aws_ssm_parameter" "bedrock_gudchangelog_agent_id" {
  name  = "/gudcommit/gudchangelog_bedrock_agent_id"
  type  = "String"
  value = aws_bedrockagent_agent.gudchangelog.agent_id
}

resource "aws_ssm_parameter" "bedrock_gudchangelog_agent_alias_id" {
  name  = "/gudcommit/gudchangelog_bedrock_agent_alias_id"
  type  = "String"
  value = aws_bedrockagent_agent_alias.gudchangelog.agent_alias_id
}
