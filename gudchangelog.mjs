import {
  BedrockAgentRuntimeClient,
  InvokeAgentCommand,
} from "@aws-sdk/client-bedrock-agent-runtime";
import {
  SSMClient,
  GetParameterCommand
} from "@aws-sdk/client-ssm"; 
import { simpleGit } from 'simple-git';

const awsAccountName = "default";
const awsRegion = "us-east-1";

const args = process.argv.slice(2);
const branchName = args[0];

if (!branchName) {
  throw new Error('No branch to compare with specified.')
}

const ssm = new SSMClient({ region: awsRegion, profile: awsAccountName });

/**
 * Retrieves the value of a parameter from AWS Systems Manager Parameter Store.
 *
 * @async
 * @param {string} name - The name of the parameter to retrieve.
 * @returns {Promise<string>} The value of the parameter.
 *
 * @example
 * const parameterValue = await getParameter('parameterName');
 * console.log(parameterValue);
 *
 * @throws Will throw an error if the AWS SSM send command fails.
 */
async function getParameter(name) {
  const input = {
    Name: name,
    WithDecryption: false,
  };
  const command = new GetParameterCommand(input);
  const response = await ssm.send(command);
  return response.Parameter.Value;
}

const agentId = await getParameter('/gudcommit/gudchangelog_bedrock_agent_id');
const agentAliasId = await getParameter('/gucommit/gudchangelog_bedrock_agent_alias_id');


/**
 * Retrieves the difference between the staged changes and the current working directory.
 * @returns {Promise<string>} A promise that resolves with the diff output as a string.
 * @throws {Error} If there's an error while getting the diff.
 */
const gitDiffResult = () => {
  return new Promise((resolve, reject) => {
    simpleGit().fetch().diff([branchName], (err, diff) => {
      if (err) {
        reject(new Error('>> Error getting diff: ', err));
      } else {
        resolve(diff);
      }
    });
  });
};


/**
 * Generates a random string of specified length.
 * @param {number} [length=8] - The desired length of the random string.
 * @returns {Promise<string>} A promise that resolves with the generated random string.
 */
const generateRandomString = async (length = 8) => {
  let result = '';
  const characters = 'ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz1234567890';

  for (let i = 0; i < length; i++) {
    result += characters.charAt(Math.floor(Math.random() * characters.length));
  }

  return result;
};

/**
 * Invokes a Bedrock agent to run an inference using the input
 * provided in the request body.
 *
 * @param {string} prompt - The prompt that you want the Agent to complete.
 * @param {string} sessionId - An arbitrary identifier for the session.
 */
export const invokeBedrockAgent = async (prompt, sessionId) => {
  const client = new BedrockAgentRuntimeClient({ region: awsRegion, profile: awsAccountName });

  const command = new InvokeAgentCommand({
    agentId,
    agentAliasId,
    sessionId,
    inputText: prompt,
  });

  try {
    let completion = "";
    const response = await client.send(command);

    if (response.completion === undefined) {
      throw new Error("Completion is undefined");
    }

    for await (let chunkEvent of response.completion) {
      const chunk = chunkEvent.chunk;
      const decodedResponse = new TextDecoder("utf-8").decode(chunk.bytes);
      completion += decodedResponse;
    }

    return { sessionId: sessionId, completion };
  } catch (err) {
    if (err.name === "ExpiredTokenException") {
      console.error(`>> The security token included in the request is expired. Re-authenticate to the ${awsAccountName} AWS account and try again.`);
    } else {
      throw err;
    }
  }
};


// Call function if run directly
const run = async () => {
  try {
    const sessionId = await generateRandomString();
    let diffOutput = await gitDiffResult();

    // Check if diffOutput is empty
    if (!diffOutput) {
      console.log('>> No changes in diff.');
      return;
    }

    diffOutput = diffOutput.replace(/\\/gm, '\\\\');
    try {
      const result = await invokeBedrockAgent(diffOutput, sessionId);
      let commitMessage;
      if (result?.completion) {
        commitMessage = result.completion.trim();
      } else {
        commitMessage = "Sorry. No changelog message could be generated.";
      }
      console.log(commitMessage);
    } catch (err) {
      throw err;
    }
  } catch (err) {
    throw err;
  }
};

try {
  run();
} catch (err) {
  console.error('>> ', err);
}
