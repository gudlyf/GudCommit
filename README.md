# GudCommit

### Generate understandable, concise git commit messages and CHANGELOG.md files based on diff's, using AWS Bedrock AI Agents.

I've been a big fan of [OpenCommit](https://github.com/di-sukharev/opencommit) for AI-generated commit messages, because sometimes you're just too time-constrained to write something more detailed than "Fixed bug." But it's currently only able to run in public AI systems like OpenAI and on local-system LLMs like [Ollama](https://ollama.com/). I wanted something that worked in my own environment, on my own terms, but works much faster and better than Ollama.

**GudCommit** runs in AWS Bedrock. The default LLM model it uses is Claude v3 (or 3.5, once that becomes available for use with agents), but it can be easily changed to use whatever model suits you.

Building this in NodeJS was the quickest for me, but it would likely be better suited as a Go application at some point.

### Usage

Clone this repository, then follow the below steps:

On Mac:

```sh
brew install awscli jq node tenv
```

Authenticate in your shell to the `default` AWS account, ensuring you have current credentials in `~/.aws/credentials` for that account and the profile is named `default`.

#### Deploy OpenTofu/Terraform

```sh
cd GudCommit/terraform/dev
tenv tofu install latest-allowed
```

Check values in `main.tf` for Terraform state information, and `variables.tf` for LLM to use. Refer to [AWS Documentation](https://docs.aws.amazon.com/bedrock/latest/userguide/model-ids.html) on the model IDs available.

```sh
tofu init
tofu apply
```

#### Update Agent Post-Processing

The `prompt_templates` are currently commented out of the Terraform code, since the provider still seems to have trouble with it. Once it is fixed, I'll uncomment the code so it can be applied without having to manually do so. Until then:

1. Navigate to the deployed Bedrock agents in the AWS console.
2. Click on each one and click the button labeled **Edit in Agent Builder**.
3. Scroll to the bottom of the page and click the **Edit** button under **Advanced prompts**.
4. Select the **Post-processing** tab. Flip both switches for **Override post-processing template defaults** and **Activate post-processing template**.
5. Find the applicable post-processing JSON in `terraform/module/prompt_templates` and paste it into the **Prompt template editor** window.
6. **Save and exit**
7. At the top of the window, click **Prepare**, then **Save and exit**.

#### Running GudCommit

Install the dependencies:

```sh
cd GudCommit && npm install
```

Create the following bash/zsh alias in your `~/.bashrc` or `./zshrc` file:

```sh
function gudco() {
    local commit_message="$(node ~/path/to/gudcommit/gudcommit.mjs)"
    echo "Generated commit message:"
    echo ""
    echo "\033[1m$commit_message\033[0m"
    echo ""
    echo -n "Proceed with the commit? (y/n or e to Edit): "
    read confirmation
    case "$confirmation" in
        [Yy])
            git commit -m "$commit_message"
            ;;
        [Ee])
            git commit -e -m "$commit_message"
            ;;
        *)
            echo "Commit canceled."
            ;;
    esac
}

function gudcl() {
    local branch=$1
    if [[ ! "$branch" ]]; then
        echo ">> Must specify a branch to compare to as argument"
        return 1
    fi
    local changelog_message="$(node ~/path/to/gudcommit/gudchangelog.mjs $branch)"
    if [[ ! "$changelog_message" ]]; then
        echo ">> No message generated."
        return 1
    fi
    echo "Generated CHANGELOG.md message:"
    echo ""
    echo "\033[1m$changelog_message\033[0m"
    echo ""
    echo -n "Prepend this content to CHANGELOG.md? (y/n): "
    read confirmation
    case "$confirmation" in
        [Yy])
            local top_level="$(git rev-parse --show-toplevel)"
            echo "$changelog_message" > /tmp/gudchangelog.md
            if [[ -f "$top_level/CHANGELOG.md" ]]; then
                cat "$top_level/CHANGELOG.md" >> /tmp/gudchangelog.md
                echo "---" >> /tmp/gudchangelog.md
            fi
            mv /tmp/gudchangelog.md "$top_level/CHANGELOG.md"
            ;;
        *)
            echo "Changelog canceled."
            ;;
    esac
}
```

```sh
source ~/.bashrc || source ~/.zshrc
```

When in another project and have added/staged files to commit (i.e. `git add .`), run the following:

```sh
gudco
```

You can also create output to be prepended to an existing or new `CHANGELOG.md` file. This will compare the current working branch with one you plan to merge into:

```sh
gudcl main
```
