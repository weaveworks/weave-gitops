import {
  ThemeProvider as MuiThemeProvider,
  StyledEngineProvider,
} from "@mui/material";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { createMemoryHistory } from "history";
import _ from "lodash";
import * as React from "react";
import { Router } from "react-router";
import { ThemeProvider } from "styled-components";
import AppContextProvider, {
  AppProps,
  ThemeTypes,
} from "../contexts/AppContext";
import { CoreClientContext } from "../contexts/CoreClientContext";
import {
  Core,
  GetChildObjectsRequest,
  GetChildObjectsResponse,
  GetReconciledObjectsRequest,
  GetReconciledObjectsResponse,
  GetVersionRequest,
  GetVersionResponse,
  ListObjectsRequest,
  ListObjectsResponse,
} from "./api/core/core.pb";
import theme, { muiTheme } from "./theme";
import { RequestError } from "./types";

export type CoreOverrides = {
  GetChildObjects?: (req: GetChildObjectsRequest) => GetChildObjectsResponse;
  GetReconciledObjects?: (
    req: GetReconciledObjectsRequest,
  ) => GetReconciledObjectsResponse;
  GetVersion?: (req: GetVersionRequest) => GetVersionResponse;
  ListObjects?: (req: ListObjectsRequest) => ListObjectsResponse;
};

export const createCoreMockClient = (
  ovr: CoreOverrides,
  error?: RequestError,
): typeof Core => {
  const promisified = _.reduce(
    ovr,
    (result, handlerFn, method) => {
      result[method] = (req) => {
        if (error) {
          return new Promise((_, reject) => reject(error));
        }
        return new Promise((accept) => accept(handlerFn(req) as any));
      };

      return result;
    },
    {},
  );

  return promisified as typeof Core;
};

export function withTheme(element, mode: ThemeTypes = ThemeTypes.Light) {
  const appliedTheme = theme(mode);
  return (
    <ThemeProvider theme={appliedTheme}>
      <StyledEngineProvider injectFirst>
        <MuiThemeProvider theme={muiTheme(appliedTheme.colors, mode)}>
          {element}
        </MuiThemeProvider>
      </StyledEngineProvider>
    </ThemeProvider>
  );
}

type TestContextProps = AppProps & {
  api?: typeof Core;
  featureFlags?: { [key: string]: string };
};

export function withContext(
  TestComponent,
  url: string,
  { api, featureFlags, ...appProps }: TestContextProps,
) {
  const history = createMemoryHistory();
  history.push(url);
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  });
  const isElement = React.isValidElement(TestComponent);
  window.matchMedia = jest.fn();
  //@ts-expect-error TODO
  window.matchMedia.mockReturnValue({ matches: false });
  return (
    <Router location={url} navigator={history}>
      <AppContextProvider footer={<></>} {...appProps}>
        <QueryClientProvider client={queryClient}>
          <CoreClientContext.Provider
            value={{ api, featureFlags: featureFlags || {} }}
          >
            {isElement ? TestComponent : <TestComponent />}
          </CoreClientContext.Provider>
        </QueryClientProvider>
      </AppContextProvider>
    </Router>
  );
}
