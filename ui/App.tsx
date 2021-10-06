import { MuiThemeProvider } from "@material-ui/core";
import qs from "query-string";
import * as React from "react";
import {
  BrowserRouter as Router,
  Redirect,
  Route,
  Switch,
} from "react-router-dom";
import { ToastContainer } from "react-toastify";
import { ThemeProvider } from "styled-components";
import ErrorBoundary from "./components/ErrorBoundary";
import Layout from "./components/Layout";
import AppContextProvider from "./contexts/AppContext";
import { Applications as appsClient } from "./lib/api/applications/applications.pb";
import theme, { GlobalStyle, muiTheme } from "./lib/theme";
import { PageRoute } from "./lib/types";
import ApplicationAdd from "./pages/ApplicationAdd";
import ApplicationDetail from "./pages/ApplicationDetail";
import Applications from "./pages/Applications";
import Error from "./pages/Error";

export default function App() {
  return (
    <MuiThemeProvider theme={muiTheme}>
      <ThemeProvider theme={theme}>
        <GlobalStyle />
        <Router>
          <AppContextProvider applicationsClient={appsClient}>
            <Layout>
              <ErrorBoundary>
                <Switch>
                  <Route
                    exact
                    path={PageRoute.Applications}
                    component={Applications}
                  />
                  <Route
                    exact
                    path={PageRoute.ApplicationDetail}
                    component={({ location }) => {
                      const params = qs.parse(location.search);

                      return <ApplicationDetail name={params.name as string} />;
                    }}
                  />
                  <Route
                    exact
                    path={PageRoute.ApplicationAdd}
                    component={ApplicationAdd}
                  />
                  <Redirect exact from="/" to={PageRoute.Applications} />
                  <Route exact path="*" component={Error} />
                </Switch>
              </ErrorBoundary>
              <ToastContainer
                position="top-center"
                autoClose={10000}
                newestOnTop={false}
              />
            </Layout>
          </AppContextProvider>
        </Router>
      </ThemeProvider>
    </MuiThemeProvider>
  );
}
