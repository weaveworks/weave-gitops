import random
import string
import requests
import json
import os

# Modify this to change the API server URL
BASE_URL = os.environ.get("BASE_URL", "http://localhost:8000")


def generate_random_string(length=8):
    return ''.join(random.choices(string.ascii_lowercase + string.digits, k=length))


def sign_in(s):
    return s.post(f"{BASE_URL}/oauth2/sign_in", json={
        "username": "wego-admin",
        "password": os.environ["WEGO_PASSWORD"]
    })


def create_kustomization():
    return {
        "metadata": {
            "name": "my-app",
        },
        "spec": {
            "path": "./apps/my-app",
            "sourceRef": {
                "name": "flux-system",
                "namespace": "flux-system"
            },
            "targetNamespace": "default",
            "createNamespace": False
        }
    }


def create_helm_release():
    return {
        "metadata": {"name": "podinfo"},
        "spec": {
            "chart": {
                "spec": {
                    "chart": "podinfo",
                    "sourceRef": {"name": "podinfo", "namespace": "prod-gitlab"},
                    "version": "6.5.3"
                }
            },
            "values": ""
        }
    }


def render(s):
    data = {
        "clusterAutomations":  [{
            "cluster": {
                "name": "management",
            },
            "kustomization": create_kustomization(),
        }, {
            "cluster": {
                "name": "management",
            },
            "helmRelease": create_helm_release(),
        }]
    }

    response = s.post(f"{BASE_URL}/v1/automations/render",
                      json=data)

    return response.json()


def create_pull_request(s):
    data = {
        "providerName": "github",
        "accessToken": os.environ["GITHUB_TOKEN"]
    }
    response = s.post(f"{BASE_URL}/v1/authenticate/github", json=data)
    token = response.json()["token"]

    headers = {
        "Git-Provider-Token": f"token {token}"
    }

    # Modify this for your repository
    org = os.environ.get("GITHUB_USER", "my-org")
    repo = os.environ.get("GITHUB_REPO", "my-repo")
    repository_url = f"https://github.com/{org}/{repo}"
    head_branch = f"branch-{generate_random_string()}"
    base_branch = "main"

    data = {
        "headBranch": head_branch,
        "clusterAutomations":  [{
            "cluster": {
                "name": "management",
            },
            "kustomization": create_kustomization(),
        }, {
            "cluster": {
                "name": "management",
            },
            "helmRelease": create_helm_release(),
        }],
        "repositoryUrl": repository_url,
        "baseBranch": base_branch
    }

    response = s.post(f"{BASE_URL}/v1/automations/pull-request",
                      json=data, headers=headers)

    return response.json()


if __name__ == "__main__":
    s = requests.Session()

    sign_in(s)

    response = render(s)

    for responseTypes in ("kustomizationFiles", "helmReleaseFiles"):
        print(f"# {responseTypes}")
        for item in response[responseTypes]:
            print(f"# {item['path']}")
            print(item["content"])

    response = create_pull_request(s)
    print(response)
