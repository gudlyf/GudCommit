Clone this repository, then follow the below steps:

```sh
brew install node
```

Authenticate in your shell to the `default` AWS account, ensuring you have current credentials in `~/.aws/credentials` for that account and the profile is named `default`.

```sh
cd gudcommit
```

```sh
npm install
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
