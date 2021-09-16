import { MuiThemeProvider } from "@material-ui/core";
import { createMemoryHistory } from "history";
import _ from "lodash";
import * as React from "react";
import { Router } from "react-router-dom";
import { ThemeProvider } from "styled-components";
import AppContextProvider from "../contexts/AppContext";
import {
  Applications,
  GetApplicationRequest,
  GetApplicationResponse,
  GetChildObjectsReq,
  GetChildObjectsRes,
  GetGithubAuthStatusRequest,
  GetGithubAuthStatusResponse,
  GetGithubDeviceCodeRequest,
  GetGithubDeviceCodeResponse,
  GetReconciledObjectsReq,
  GetReconciledObjectsRes,
  ListApplicationsRequest,
  ListApplicationsResponse,
  ListCommitsRequest,
  ListCommitsResponse,
} from "./api/applications/applications.pb";
import theme, { muiTheme } from "./theme";

export type ApplicationOverrides = {
  ListApplications?: (req: ListApplicationsRequest) => ListApplicationsResponse;
  GetApplication?: (req: GetApplicationRequest) => GetApplicationResponse;
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
};

// Don't make the user wire up all the promise stuff to be interface-compliant
export const createMockClient = (
  ovr: ApplicationOverrides
): typeof Applications => {
  const promisified = _.reduce(
    ovr,
    (result, handlerFn, method) => {
      result[method] = (req) =>
        new Promise((accept) => accept(handlerFn(req) as any));

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

export function withContext(
  TestComponent,
  url: string,
  appOverrides?: ApplicationOverrides
) {
  const history = createMemoryHistory();
  history.push(url);

  const isElement = React.isValidElement(TestComponent);
  return (
    <Router history={history}>
      <AppContextProvider
        applicationsClient={createMockClient(appOverrides) as any}
      >
        {isElement ? TestComponent : <TestComponent />}
      </AppContextProvider>
    </Router>
  );
}
