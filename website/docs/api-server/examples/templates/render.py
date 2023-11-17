import requests
import json
import os

# Modify this to change the API server URL
BASE_URL = os.environ.get("BASE_URL", "http://localhost:8000")


def sign_in(s):
    return s.post(f"{BASE_URL}/oauth2/sign_in", json={
        "username": "wego-admin",
        "password": os.environ["WEGO_PASSWORD"]
    })


def render(s):
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

    response = s.post(f"{BASE_URL}/v1/namespaces/{namespace}/templates/{name}/render",
                      json=data)

    return response.json()


if __name__ == "__main__":
    s = requests.Session()

    sign_in(s)

    response = render(s)

    for item in response["renderedTemplates"]:
        print(f"# {item['path']}")
        print(item["content"])
