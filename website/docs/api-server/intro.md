---
title: Introduction
---

Some use cases require programmatic access to Weave GitOps. For example integrating Weave GitOps with other systems or automating GitOps workflows. The Weave GitOps API server provides a REST API for this. The UI is built on top of this API.

## Accessing the Weave GitOps API server

The Weave GitOps API server can be accessed similarly to how the UI accesses it - by making HTTP requests to the API server's URL.
If you've exposed the UI via an Ingress endpoint that is accessible from other systems, the API will also be available at that same URL path.
If you've not exposed the UI via an Ingress, port-forwarding can be used to access it locally.

## Quickstart

Lets go through a few simple examples of how to use the API server by listing `HelmRelease` objects visible to the user.

### Using curl

The session system uses cookies, so we'll need to login first to get a session cookie:

Login with JSON payload and save the session cookies to `cookies.txt`.

```bash
curl \
  --fail \
  --cookie-jar cookies.txt \
  --data '{"username":"your_username","password":"your_password"}' \
  http://localhost:8000/oauth2/sign_in || "Login failed"
```

:::tip

Add curl's `-v` flag to the command see the HTTP request and response headers to debug login and other issues.

:::


Use the saved session cookie to list `HelmRelease` objects visible to the user:

```bash
curl \
  --cookie cookies.txt \
  --data '{ "kind": "HelmRelease" }' http://localhost:8000/v1/objects
```

### Using Node.js

Here's an example using node.js to login and list `HelmRelease` objects visible to the user:

:::note

This example uses `fetch` which is available from node v18.0.0. If you're using an older version of node, you can use a library like [node-fetch](https://www.npmjs.com/package/node-fetch) instead.

:::

```js title="list-helmreleases.mjs"
const response = await fetch("http://localhost:8000/oauth2/sign_in", {
  method: "POST",
  body: JSON.stringify({
    username: "your_username",
    password: "your_password",
  }),
});

if (!response.ok) {
  const error = await response.text();
  throw new Error(`Request failed with status ${response.status}: ${error}`);
}

const cookie = (await response).headers.get("set-cookie");

// Get objects
const objectsResponse = await fetch("http://localhost:8000/v1/objects", {
  method: "POST",
  headers: { Cookie: cookie },
  body: JSON.stringify({
    kind: "HelmRelease",
  }),
});

const data = await objectsResponse.json();

console.log(data);
```

### Using Python

Here's an example using Python to login and list `HelmRelease` objects visible to the user:

```python title="list-helmreleases.py"
import requests

# Login
response = requests.post(
    "http://localhost:8000/oauth2/sign_in",
    json={"username": "your_username", "password": "your_password"},
)

if not response.ok:
    raise Exception(
        f"Request failed with status {response.status_code}: {response.text}"
    )

cookie = response.headers["set-cookie"]

# Get objects
objects_response = requests.post(
    "http://localhost:8000/v1/objects",
    json={"kind": "HelmRelease"},
    headers={"Cookie": cookie},
)

print(objects_response.json())

```

## Notes on the API server

### Some methods use POST rather than GET for querying

Some endpoints use POST rather than GET for querying.
Much of the server request layer is auto-generated, complex query parameters are not supported by the auto-generated code, so POST is used instead.