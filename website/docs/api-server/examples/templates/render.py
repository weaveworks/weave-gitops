import requests
import json
import os

# Modify this to change the API server URL
BASE_URL = os.environ.get("BASE_URL", "http://localhost:8000")
WEGO_PASSWORD = os.environ["WEGO_PASSWORD"]

session_cookie = ""
headers_default = {
    "Content-Type": "application/json"
}


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


def render():
    # Modify this for the desired template
    namespace = "default"
    name = "vcluster-template-development"

    # Modify or add more parameters as needed
    parameter_values = {
        "CLUSTER_NAME": "foo",
    }

    data = {
        "parameterValues": parameter_values,
        "templateKind": "GitOpsTemplate",
    }

    response = request(
        f"/v1/namespaces/{namespace}/templates/{name}/render", "POST", data)

    for item in response["renderedTemplates"]:
        print("#")
        print(f"# {item['path']}")
        print("#")
        print(item["content"])


if __name__ == "__main__":
    data = {
        "username": "wego-admin",
        "password": WEGO_PASSWORD
    }
    request("/oauth2/sign_in", "POST", data)
    render()
