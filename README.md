# Stargazers Insights

A small CLI tool to export insights about stargazers of a GitHub repository.

## Prerequisites

1. **Create a GitHub token**:
   To access the GitHub API, you need to create a personal access token. Follow the instructions in the [GitHub documentation](https://docs.github.com/en/authentication/keeping-your-account-and-data-secure/creating-a-personal-access-token) to generate your token.

2. **Create a `stargazers.yaml` file**:
   This file should contain the repositories you want to analyze. Use the following structure:

   ```yaml
   repositories:
     - owner: openstatusHQ
       name: cli
     - owner: openstatusHQ
       name: openstatus
    ```

## Contributing
Contributions are welcome! Please open an issue or submit a pull request for any features or bug fixes.
