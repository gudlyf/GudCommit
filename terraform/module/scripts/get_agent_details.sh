#!/usr/bin/env bash

AGENT_ID=$1
AGENT_VERSION=$(aws --profile default bedrock-agent list-agents --query "agentSummaries[?agentId=='${AGENT_ID}'].latestAgentVersion" --output text)

jq -n --arg agent_version "${AGENT_VERSION}" '{"agent_version":$agent_version}'
