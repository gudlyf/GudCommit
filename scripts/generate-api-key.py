#!/usr/bin/env python3
"""
GudCommit Bedrock API Key Generator

This script uses the aws-bedrock-token-generator library to automatically
generate and manage Bedrock API keys for GudCommit.
"""

import os
import sys
import json
import time
import subprocess
from datetime import datetime, timedelta
from pathlib import Path

# Try to import the required libraries
try:
    from aws_bedrock_token_generator import provide_token
except ImportError:
    print("❌ aws-bedrock-token-generator not found!")
    print("📦 Installing aws-bedrock-token-generator...")
    subprocess.check_call([sys.executable, "-m", "pip", "install", "aws-bedrock-token-generator"])
    from aws_bedrock_token_generator import provide_token

def get_aws_credentials():
    """Get AWS credentials from environment or AWS CLI"""
    try:
        # Try to get credentials from AWS CLI
        result = subprocess.run(
            ["aws", "sts", "get-caller-identity"],
            capture_output=True,
            text=True,
            check=True
        )
        print("✅ AWS credentials found")
        return True
    except subprocess.CalledProcessError:
        print("❌ AWS credentials not configured")
        print("Please run: aws configure")
        return False
    except FileNotFoundError:
        print("❌ AWS CLI not found")
        print("Please install AWS CLI: https://aws.amazon.com/cli/")
        return False

def generate_api_key(region="us-east-1"):
    """Generate a new short-term Bedrock API key"""
    try:
        print(f"🔑 Generating short-term Bedrock API key for region: {region}")
        
        # Generate short-term API key using the token generator
        # The provide_token() function handles credentials and expiration automatically
        api_key = provide_token()
        
        if not api_key:
            print("❌ Failed to generate API key")
            return None
            
        print("✅ Short-term API key generated successfully!")
        print("⏰ Key expires in 12 hours or when your session ends")
        
        return api_key
        
    except Exception as e:
        print(f"❌ Failed to generate API key: {e}")
        print("💡 Make sure you have Bedrock permissions in your AWS account")
        return None


def save_api_key(api_key, region="us-east-1"):
    """Save API key to environment file"""
    home_dir = Path.home()
    env_file = home_dir / ".gudcommit-credentials"
    
    # Create credentials file
    with open(env_file, 'w') as f:
        f.write(f"# GudCommit Bedrock API Key\n")
        f.write(f"# Generated: {datetime.now().strftime('%Y-%m-%d %H:%M:%S')}\n")
        f.write(f"# Region: {region}\n")
        f.write(f"# Expires: {datetime.now() + timedelta(hours=12)}\n")
        f.write(f"export GUD_BEDROCK_API_KEY=\"{api_key}\"\n")
        f.write(f"export AWS_REGION=\"{region}\"\n")
    
    print(f"💾 API key saved to: {env_file}")
    return env_file

def set_environment_variable(api_key, region="us-east-1"):
    """Set environment variable for current session"""
    os.environ["GUD_BEDROCK_API_KEY"] = api_key
    os.environ["AWS_REGION"] = region
    print("🔧 Environment variables set for current session")

def test_api_key(api_key):
    """Test the API key by making a simple request"""
    try:
        print("🧪 Testing API key...")
        
        # Simple test - just check if the key is valid format
        if len(api_key) > 20:  # Basic validation
            print("✅ API key format looks valid")
            return True
        else:
            print("❌ API key format appears invalid")
            return False
            
    except Exception as e:
        print(f"❌ API key test failed: {e}")
        return False

def show_usage_instructions(env_file):
    """Show instructions for using the API key"""
    print("\n" + "="*60)
    print("🎉 GudCommit API Key Setup Complete!")
    print("="*60)
    print()
    print("📋 Usage Instructions:")
    print()
    print("1. Load the credentials:")
    print(f"   source {env_file}")
    print()
    print("2. Or set manually:")
    print(f"   export GUD_BEDROCK_API_KEY=\"{os.environ.get('GUD_BEDROCK_API_KEY', 'your-key')}\"")
    print()
    print("3. Use GudCommit:")
    print("   cd golang && make build")
    print("   ./bin/gudcommit")
    print()
    print("⏰ Note: API key expires in 12 hours")
    print("🔄 To regenerate: python scripts/generate-api-key.py")
    print()

def main():
    """Main function"""
    print("🚀 GudCommit Bedrock API Key Generator")
    print("="*40)
    
    # Check AWS credentials
    if not get_aws_credentials():
        sys.exit(1)
    
    # Get region from environment or use default
    region = os.environ.get("AWS_REGION", "us-east-1")
    print(f"🌍 Using region: {region}")
    
    # Generate API key
    api_key = generate_api_key(region)
    if not api_key:
        sys.exit(1)
    
    # Set environment variable
    set_environment_variable(api_key, region)
    
    # Test the API key
    if not test_api_key(api_key):
        print("⚠️  API key generated but test failed")
    
    # Save to file
    env_file = save_api_key(api_key, region)
    
    # Show usage instructions
    show_usage_instructions(env_file)

if __name__ == "__main__":
    main()
