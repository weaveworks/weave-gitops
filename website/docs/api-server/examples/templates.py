import requests
import json
import sys
import os
import argparse
import random
import string
from typing import Dict, Any, Optional, Union

session_cookie = ""
headers_default: Dict[str, str] = {
    "Content-Type": "application/json"
}


def generate_random_string(length: int = 8) -> str:
    return ''.join(random.choices(string.ascii_lowercase + string.digits, k=length))


def wego_request(base_url: str, path: str, method: str = "GET", data: Optional[Dict[str, Any]] = None, headers: Optional[Dict[str, str]] = None) -> Any:
    global session_cookie
    url = base_url + path
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


def list_templates(base_url: str) -> None:
    response = wego_request(base_url, "/v1/templates")
    for obj in response["templates"]:
        namespace = obj["namespace"]
        name = obj["name"]
        template_kind = obj["templateKind"]
        required_params_names = [
            p["name"] for p in obj["parameters"] if p["required"] and not p["default"]]
        print(
            f"{namespace}/{name} template_kind={template_kind} params={required_params_names}")


def create_pull_request(args: argparse.Namespace) -> None:
    data = {
        "providerName": "github",
        "accessToken": os.environ["GITHUB_TOKEN"]
    }
    response = wego_request(
        args.base_url, "/v1/authenticate/github", "POST", data)
    token = response["token"]

    template_namespace, template_name = args.template.split('/')

    parameter_values = {}
    if args.parameter_values:
        for param in args.parameter_values:
            key, *value = param.split('=')
            # Join the value parts back together in case they contain '='
            value = '='.join(value)
            parameter_values[key] = value

    data = {
        "headBranch": f"{args.head_branch}-{generate_random_string()}",
        "templateName": template_name,
        "templateNamespace": template_namespace,
        "parameterValues": parameter_values,
        "templateKind": "GitOpsTemplate",
        "repositoryUrl": args.repository_url,
        "baseBranch": args.base_branch
    }
    headers = {
        "Git-Provider-Token": f"token {token}"
    }
    response = wego_request(
        args.base_url, "/v1/templates/pull-request", "POST", data, headers)
    print(response["webUrl"])


def main() -> None:
    parser = argparse.ArgumentParser(
        description="CLI tool for managing templates and pull requests.")
    parser.add_argument(
        "--base-url", default="http://localhost:8000", help="Base URL for the server.")
    subparsers = parser.add_subparsers(dest="command")

    # list_templates command
    subparsers.add_parser("ls", help="List templates.")

    # create_pull_request command
    pr_parser = subparsers.add_parser(
        "create-pr", help="Create a pull request.")
    pr_parser.add_argument("--head-branch", default="render-template",
                           help="Prefix for the head branch for the pull request. A random suffix will be added to avoid collisions.")
    pr_parser.add_argument("--template", required=True,
                           help="Template in the format namespace/name.")
    pr_parser.add_argument("--repository-url", required=True,
                           help="URL of the repository.")
    pr_parser.add_argument("--base-branch", default="main",
                           help="Base branch for the pull request.")
    pr_parser.add_argument("--parameter-values", nargs="*",
                           help="Parameter values in the format key=value.")

    args = parser.parse_args()

    if not args.command:
        parser.print_help()
        sys.exit(1)

    if "WEGO_PASSWORD" not in os.environ:
        raise ValueError("WEGO_PASSWORD env var is required")

    data = {
        "username": "wego-admin",
        "password": os.environ["WEGO_PASSWORD"]
    }
    wego_request(args.base_url, "/oauth2/sign_in", "POST", data)

    if args.command == "ls":
        list_templates(args.base_url)
    elif args.command == "create-pr":
        create_pull_request(args)


if __name__ == "__main__":
    main()
