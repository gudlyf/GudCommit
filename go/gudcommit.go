package main

import (
    "context"
    "crypto/rand"
    "encoding/base64"
    "fmt"
    "log"

    "github.com/aws/aws-sdk-go-v2/aws"
    "github.com/aws/aws-sdk-go-v2/config"
    "github.com/aws/aws-sdk-go-v2/service/ssm"
    "github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
    bedrockruntimetypes "github.com/aws/aws-sdk-go-v2/service/bedrockruntime/types"
    "github.com/go-git/go-git/v5"
)

const (
    awsRegion    = "us-east-1"
    awsAccountName = "default"
)

func getParameter(ctx context.Context, client *ssm.Client, name string) (string, error) {
    input := &ssm.GetParameterInput{
        Name:           aws.String(name),
        WithDecryption: aws.Bool(false),
    }
    result, err := client.GetParameter(ctx, input)
    if err != nil {
        return "", err
    }
    return *result.Parameter.Value, nil
}

func gitDiff() (string, error) {
    repo, err := git.PlainOpen(".")
    if err != nil {
        return "", err
    }
    wt, err := repo.Worktree()
    if err != nil {
        return "", err
    }
    status, err := wt.Status()
    if err != nil {
        return "", err
    }
    return status.String(), nil
}

func generateRandomString(length int) (string, error) {
    b := make([]byte, length)
    _, err := rand.Read(b)
    if err != nil {
        return "", err
    }
    return base64.URLEncoding.EncodeToString(b)[:length], nil
}

func invokeBedrockAgent(ctx context.Context, client *bedrockruntime.Client, agentId, agentAliasId, prompt, sessionId string) (string, error) {
    input := &bedrockruntimetypes.InvokeAgentInput{
        AgentId:       aws.String(agentId),
        AgentAliasId:  aws.String(agentAliasId),
        SessionId:     aws.String(sessionId),
        InputText:     aws.String(prompt),
    }
    result, err := client.InvokeAgent(ctx, input)
    if err != nil {
        return "", err
    }
    // Handling the event stream
    eventStream := result.GetStream()
    defer eventStream.Close()

    var completion string
    for event := range eventStream.Events() {
        if event.Err != nil {
            return "", event.Err
        }
        if msg, ok := event.Response.(*bedrockruntimetypes.InvokeAgentResponseStream_Chunk); ok {
            completion += *msg.Chunk
        }
    }
    if err := eventStream.Err(); err != nil {
        return "", err
    }
    return completion, nil
}

func main() {
    ctx := context.TODO()
    cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(awsRegion))
    if (err != nil) {
        log.Fatalf("unable to load SDK config, %v", err)
    }

    ssmClient := ssm.NewFromConfig(cfg)

    agentId, err := getParameter(ctx, ssmClient, "/gudcommit/gudcommit_bedrock_agent_id")
    if (err != nil) {
        log.Fatalf("failed to get parameter: %v", err)
    }

    agentAliasId, err := getParameter(ctx, ssmClient, "/gudcommit/gudcommit_bedrock_agent_alias_id")
    if (err != nil) {
        log.Fatalf("failed to get parameter: %v", err)
    }

    diffOutput, err := gitDiff()
    if (err != nil) {
        log.Fatalf("failed to get git diff: %v", err)
    }

    if (diffOutput == "") {
        fmt.Println(">> No changes to commit.")
        return
    }

    sessionId, err := generateRandomString(8)
    if (err != nil) {
        log.Fatalf("failed to generate session ID: %v", err)
    }

    bedrockClient := bedrockruntime.NewFromConfig(cfg)

    result, err := invokeBedrockAgent(ctx, bedrockClient, agentId, agentAliasId, diffOutput, sessionId)
    if (err != nil) {
        log.Fatalf("failed to invoke Bedrock agent: %v", err)
    }

    fmt.Println(result)
}
