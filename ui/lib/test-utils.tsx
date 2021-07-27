import { MuiThemeProvider } from "@material-ui/core";
import { createMemoryHistory } from "history";
import _ from "lodash";
import * as React from "react";
import { Router } from "react-router-dom";
import { ThemeProvider } from "styled-components";
import AppContextProvider from "../contexts/AppContext";
import {
  GetApplicationResponse,
  ListApplicationsResponse,
} from "./api/applications/applications.pb";
import theme, { muiTheme } from "./theme";

type ApplicationOverrides = {
  ListApplications?: ListApplicationsResponse;
  GetApplication?: GetApplicationResponse;
};

// Don't make the user wire up all the promise stuff to be interface-compliant
export const createMockClient = (ovr: ApplicationOverrides) => {
  const promisified = _.reduce(
    ovr,
    (result, desiredResponse, method) => {
      result[method] = () =>
        new Promise((accept) => accept(desiredResponse as any));

      return result;
    },
    {}
  );

  return promisified;
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
  return (
    <Router history={history}>
      <AppContextProvider applicationsClient={createMockClient(appOverrides)}>
        <TestComponent />
      </AppContextProvider>
    </Router>
  );
}
