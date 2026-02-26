#!/bin/bash
# Helper script to list available Bedrock models and inference profiles

set -e

REGION="${AWS_REGION:-us-east-1}"
PROFILE="${AWS_PROFILE:-gudcommit}"

echo "========================================="
echo "Bedrock Models & Inference Profiles"
echo "========================================="
echo "Region: $REGION"
echo "Profile: $PROFILE"
echo ""

echo "--- Direct Invoke Models (ON_DEMAND) ---"
echo "These models can be used directly with model_id:"
echo ""
aws bedrock list-foundation-models \
  --region "$REGION" \
  --profile "$PROFILE" \
  --query "modelSummaries[?contains(inferenceTypesSupported, 'ON_DEMAND')].{ModelId:modelId,Provider:providerName,Status:modelLifecycle.status}" \
  --output table 2>/dev/null || echo "Error: Could not list models. Check your credentials."

echo ""
echo "--- Inference Profiles ---"
echo "Newer Claude models require using inference profile ARNs:"
echo ""
aws bedrock list-inference-profiles \
  --region "$REGION" \
  --profile "$PROFILE" \
  --query 'inferenceProfileSummaries[].{Name:inferenceProfileName,ARN:inferenceProfileArn,Type:type,Models:models[0].modelId}' \
  --output table 2>/dev/null || echo "Error: Could not list inference profiles. Check your credentials."

echo ""
echo "========================================="
echo "Usage Examples:"
echo "========================================="
echo ""
echo "1. For older models (works with direct model ID):"
echo '   model_id: "anthropic.claude-3-5-sonnet-20240620-v1:0"'
echo ""
echo "2. For newer models (requires inference profile ARN):"
echo '   Find the ARN from the list above and use it as model_id'
echo '   Example: "arn:aws:bedrock:us-east-1::foundation-model/anthropic.claude-sonnet-4-5-v1:0"'
echo ""
echo "3. Add to ~/.gudcommit.json:"
echo "   {"
echo '     "model_id": "your-model-id-or-arn-here",'
echo '     "timeout_seconds": 60,'
echo '     "region": "us-east-1"'
echo "   }"
echo ""
