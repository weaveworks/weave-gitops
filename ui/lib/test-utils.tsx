import { MuiThemeProvider } from "@material-ui/core";
import { createMemoryHistory } from "history";
import _ from "lodash";
import * as React from "react";
import { Router } from "react-router-dom";
import { ThemeProvider } from "styled-components";
import AppContextProvider, { AppProps } from "../contexts/AppContext";
import {
  Applications,
  GetChildObjectsReq,
  GetChildObjectsRes,
  GetGithubAuthStatusRequest,
  GetGithubAuthStatusResponse,
  GetGithubDeviceCodeRequest,
  GetGithubDeviceCodeResponse,
  GetReconciledObjectsReq,
  GetReconciledObjectsRes,
  ListCommitsRequest,
  ListCommitsResponse,
  ParseRepoURLRequest,
  ParseRepoURLResponse,
  SyncApplicationRequest,
  SyncApplicationResponse,
} from "./api/applications/applications.pb";
import theme, { muiTheme } from "./theme";
import { RequestError } from "./types";

export type ApplicationOverrides = {
  ListCommits?: (req: ListCommitsRequest) => ListCommitsResponse;
  GetReconciledObjects?: (
    req: GetReconciledObjectsReq
  ) => GetReconciledObjectsRes;
  GetChildObjects?: (req: GetChildObjectsReq) => GetChildObjectsRes;
  GetGithubDeviceCode?: (
    req: GetGithubDeviceCodeRequest
  ) => GetGithubDeviceCodeResponse;
  GetGithubAuthStatus?: (
    req: GetGithubAuthStatusRequest
  ) => GetGithubAuthStatusResponse;
  ParseRepoURL?: (req: ParseRepoURLRequest) => ParseRepoURLResponse;
  SyncApplication?: (req: SyncApplicationRequest) => SyncApplicationResponse;
};

// Don't make the user wire up all the promise stuff to be interface-compliant
export const createMockClient = (
  ovr: ApplicationOverrides,
  error?: RequestError
): typeof Applications => {
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
    {}
  );

  return promisified as typeof Applications;
};

export function withTheme(element) {
  return (
    <MuiThemeProvider theme={muiTheme}>
      <ThemeProvider theme={theme}>{element}</ThemeProvider>
    </MuiThemeProvider>
  );
}

export function withContext(TestComponent, url: string, ctxProps: AppProps) {
  const history = createMemoryHistory();
  history.push(url);

  const isElement = React.isValidElement(TestComponent);
  return (
    <Router history={history}>
      <AppContextProvider renderFooter {...ctxProps}>
        {isElement ? TestComponent : <TestComponent />}
      </AppContextProvider>
    </Router>
  );
}
