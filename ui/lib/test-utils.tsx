import { MuiThemeProvider } from "@material-ui/core";
import { createMemoryHistory } from "history";
import * as React from "react";
import { Router } from "react-router-dom";
import { ThemeProvider } from "styled-components";
import theme, { muiTheme } from "./theme";

export function withTheme(element) {
  return (
    <MuiThemeProvider theme={muiTheme}>
      <ThemeProvider theme={theme}>{element}</ThemeProvider>
    </MuiThemeProvider>
  );
}

export function withContext(TestComponent, url: string) {
  const history = createMemoryHistory();
  history.push(url);
  return (
    <Router history={history}>
      <TestComponent />
    </Router>
  );
}
