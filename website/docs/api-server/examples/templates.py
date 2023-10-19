import requests
import json
import os
import random
import string

BASE_URL = "http://localhost:8000"  # Modify this to change the API server URL
GITHUB_TOKEN = os.environ["GITHUB_TOKEN"]
WEGO_PASSWORD = os.environ["WEGO_PASSWORD"]

session_cookie = ""
headers_default = {
    "Content-Type": "application/json"
}


def generate_random_string(length=8):
    return ''.join(random.choices(string.ascii_lowercase + string.digits, k=length))


def request(path, method="GET", data=None, headers=None):
    global session_cookie
    url = BASE_URL + path
    headers_merged = {**headers_default, **(headers or {})}
    if session_cookie:
        headers_merged["Cookie"] = session_cookie

    response = requests.request(
        method, url, data=json.dumps(data), headers=headers_merged)

    if not response.ok:
        raise ValueError(
            f"Request {url} failed with status {response.status_code}: {response.text}")

    if "set-cookie" in response.headers:
        session_cookie = response.headers["set-cookie"]

    if not response.text:
        return None

    return response.json()


def create_pull_request():
    data = {
        "providerName": "github",
        "accessToken": GITHUB_TOKEN
    }
    response = request("/v1/authenticate/github", "POST", data)
    token = response["token"]

    # Define pull request data
    head_branch = f"branch-{generate_random_string()}"
    # Modify this for the desired template
    template = "default/vcluster-template-development"
    template_namespace, template_name = template.split('/')
    # Modify this for your repository
    repository_url = "https://github.com/my-org/my-repo"
    base_branch = "main"
    parameter_values = {
        "CLUSTER_NAME": "foo",  # Modify or add more parameters as needed
    }

    data = {
        "headBranch": head_branch,
        "templateName": template_name,
        "templateNamespace": template_namespace,
        "parameterValues": parameter_values,
        "templateKind": "GitOpsTemplate",
        "repositoryUrl": repository_url,
        "baseBranch": base_branch
    }
    headers = {
        "Git-Provider-Token": f"token {token}"
    }

    response = request("/v1/templates/pull-request", "POST", data, headers)
    print(response["webUrl"])


if __name__ == "__main__":
    data = {
        "username": "wego-admin",
        "password": WEGO_PASSWORD
    }
    request("/oauth2/sign_in", "POST", data)
    create_pull_request()
