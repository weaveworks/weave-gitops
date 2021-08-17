# 7. Embedding Weave GitOps Core

Date: 2021-08-12

## Status

Accepted

## Problem

We want to be able to use the Weave GitOps Core UI elements and support services in other projects. These projects might be other editions of Weave GitOps, or other third-party environments.

Weave GitOps will need to be designed in such as way as to support this type of installation, and provide the necessary facilities to make embedding the UI as simple as possible.

## Design

### WeGo Core UI as a library

The Weave GitOps Core repository will contain an `index.js` file that will define the public API for including WeGO UI components in other projects. That file might look something like this:

```javascript
import _Timestamp from "./components/Timestamp.tsx";
import _AppContextProvider from "./contexts/AppContext";
import { Applications as appsClient } from "./lib/api/applications/applications.pb";
import _Theme from "./lib/theme";
import _ApplicationDetail from "./pages/ApplicationDetail";
import _Applications from "./pages/Applications";

export const Timestamp = _Timestamp;
export const theme = _Theme;
export const AppContextProvider = _AppContextProvider;
export const Applications = _Applications;
export const ApplicationDetail = _ApplicationDetail;
export const applicationsClient = appsClient;
```

From this file, a library bundle will be generated. This bundle will contain transpiled code that can be directly imported into another project. The contents of this bundle are not meant to be human-readable, as it will be translated from TypeScript into a browser-friendly version of JS. Additionally, the bundle will also contain a typescript definitions file so that consumers of this bundle can have a smoother developer experience.

This bundle will be stored as a public GitHub Package. This will allow us to associate a package version with a source code release. These packages will be built during CI and ONLY as part of a Weave GitOps release.

We are making a deliberate choice NOT to use the global NPM registry, as it would require a separate authentication step with NPM.org in order to push new versions of the package. Installing an NPM package from GitHub packages will require users to create an `.npmrc` file with the following entry (from the GitHub packages docs):

```
@weaveworks:registry=https://npm.pkg.github.com
```

This will associate the `@weaveworks` organization with the NPM organization specified in the package.json.

Then to install the package into a project:

```json
{
  "name": "@my-org/some-project",
  "dependencies": {
    "@weaveworks/weave-gitops": "1.0.0"
  }
}
```

There will be a specific order in which some components will need to be wrapped in order for them to work properly. For example, the `<Applications />` component expects a specifc `React.Context` to be available. The WeGO UI package will provide the necessary supporting modules:

```typescript
import * as React from "react";
import { ThemeProvider } from "styled-components";
import {
  AppContextProvider,
  ApplicationDetail,
  applicationsClient,
  theme,
} from "@weaveworks/weave-gitops";

export default function App() {
  return (
    <div>
      <ThemeProvider theme={theme}>
        <AppContextProvider
          linkResolver={linkResolver}
          applicationsClient={applicationsClient}
        >
          <ApplicationDetail />
        </AppContextProvider>
      </ThemeProvider>
    </div>
  );
}
```

This allows for the components to be embedded anywhere in the consuming application, as well as allowing the WeGO components to remain extensible and loosely-coupled to the consumer.

Consumers will need to provide a `linkResolver` function to handle navigation and routing differences in their respective environment:

```typescript
const APPS_ROUTE = "/my_custom_applications_route";
const APP_DETAIL_ROTUE = "/my_special_application_detail_route";

// Use this to reconcile differences in routing so that links will work correctly.
const linkResolver = (incoming: string): string => {
  const parsed = new URL(incoming, window.location.href);

  switch (parsed.pathname) {
    case "/applications":
      return `${APPS_ROUTE}${parsed.search}`;

    case "/application_detail":
      return `${APP_DETAIL_ROTUE}${parsed.search}`;

    default:
      return incoming;
  }
};

function MyApp() {
  return (
    <AppContextProvider
      linkResolver={linkResolver} //Link resolver added to context
      applicationsClient={applicationsClient}
    >
      <ApplicationDetail />
    </AppContextProvider>
  );
}
```

For local development, we can utilize the `npm link` functionality to edit both project simultaneously and see the results in the consuming project.

**Note** The public API for WeGO components will be `props` and `context`. WeGO components will NOT rely on URL parameters directly.

### WeGO Core HTTP Services

Certain WeGO Core components will rely on HTTP backends being available. WeGO Core will export the necessary calls to serve these backends via the default `go` module system. An example of an export HTTP service:

```golang

func NewAppsHTTPHandler(ctx context.Context, opts ...runtime.ServeMuxOption) (http.Handler, error) {
	mux := runtime.NewServeMux(opts...)

	if err := pb.RegisterApplicationsHandlerServer(ctx, mux, NewApplicationsServer()); err != nil {
		return nil, fmt.Errorf("could not register application: %w", err)
	}

    // Returns a builtin go http handler
	return mux, nil
}
```

Then to consume this function:

```golang

package main

import (
	"net/http"

	"github.com/weaveworks/weave-gitops/pkg/server"
)

var addr = "0.0.0.0:8000"

func main() {
	mux := http.NewServeMux()

	appsHandler, err := server.NewAppsHTTPHandler()
	if err != nil {
		panic(err)
	}

	mux.Handle("/apps/v1", appsHandler)
	mux.Handle("/some_other_route", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.Write([]byte("hello!"))
	}))

	if err := http.ListenAndServe(addr, mux); err != nil {
		panic(err)
	}

}
```

The WeGO Core components will then make their API requests to the `/apps/v1/*` route to fetch the data they need.

## Decision

Implement the functionality described in the Design section.

## Alternatives Considered

- `<iframe />` embedding: We decided against this because it would isolate the two applications and not provide no extensibility for the consumer of the WeGO package
- Run-time lazy loading: We decided against this because it would still require the server to have the correct UI embedded at build time, plus the added complexity and percieved performance impact.

## Consequences

Going forward, sharing functionality and code between the different editions will allow both projects to move more quickly and avoid duplication of effort. Modelling everything as a build-time dependency should help highlight problems earlier in the development cycle.

A general risk factor of this approach is that both projects much maintain compatible version of shared dependencies. For example, if both projects use the `controller-runtime` package, the `go` module system will require that those depedencies resolve to a single compatible version (which may case build errors). JS package managers are more flexible and will import two versions of the same depdency to satisfy both import cases.
