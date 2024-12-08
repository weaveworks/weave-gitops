import requests
import json
import os
import random
import string

# Modify this to change the API server URL
BASE_URL = os.environ.get("BASE_URL", "http://localhost:8000")


def sign_in(s):
    return s.post(f"{BASE_URL}/oauth2/sign_in", json={
        "username": "wego-admin",
        "password": os.environ["WEGO_PASSWORD"]
    })


def generate_random_string(length=8):
    return ''.join(random.choices(string.ascii_lowercase + string.digits, k=length))


def create_pull_request(s):
    data = {
        "providerName": "github",
        "accessToken": os.environ["GITHUB_TOKEN"]
    }
    response = s.post(f"{BASE_URL}/v1/authenticate/github", json=data)
    token = response.json()["token"]

    # Modify this for your repository
    org = os.environ.get("GITHUB_USER", "my-org")
    repo = os.environ.get("GITHUB_REPO", "my-repo")
    repository_url = f"https://github.com/{org}/{repo}"
    head_branch = f"branch-{generate_random_string()}"
    base_branch = "main"

    # Modify this for the desired template
    namespace = "default"
    name = "vcluster-template-development"

    # Modify or add more parameters as needed
    parameter_values = {
        "CLUSTER_NAME": "foo",
    }

    data = {
        "headBranch": head_branch,
        "parameterValues": parameter_values,
        "templateKind": "GitOpsTemplate",
        "repositoryUrl": repository_url,
        "baseBranch": base_branch
    }

    headers = {
        "Git-Provider-Token": f"token {token}"
    }

    response = s.post(f"{BASE_URL}/v1/namespaces/{namespace}/templates/{name}/pull-request",
                      json=data, headers=headers)

    return response.json()


if __name__ == "__main__":
    s = requests.Session()

    sign_in(s)

    response = create_pull_request(s)

    print(response["webUrl"])
