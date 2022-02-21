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
import "react-toastify/dist/ReactToastify.css";
import { ThemeProvider } from "styled-components";
import ErrorBoundary from "./components/ErrorBoundary";
import Layout from "./components/Layout";
import AppContextProvider from "./contexts/AppContext";
import AuthContextProvider, { AuthCheck } from "./contexts/AuthContext";
import {
  Applications as appsClient,
  GitProvider,
} from "./lib/api/applications/applications.pb";
import Fonts from "./lib/fonts";
import theme, { GlobalStyle, muiTheme } from "./lib/theme";
import { PageRoute } from "./lib/types";
import ApplicationAdd from "./pages/ApplicationAdd";
import ApplicationDetail from "./pages/ApplicationDetail";
import ApplicationRemove from "./pages/ApplicationRemove";
import Applications from "./pages/Applications";
import Error from "./pages/Error";
import OAuthCallback from "./pages/OAuthCallback";
import SignIn from "./pages/SignIn";

export default function App() {
  return (
    <MuiThemeProvider theme={muiTheme}>
      <ThemeProvider theme={theme}>
        <Fonts />
        <GlobalStyle />
        <Router>
          <AuthContextProvider>
            <Switch>
              {/* <Signin> does not use the base page <Layout> so pull it up here */}
              <Route component={SignIn} exact={true} path="/sign_in" />
              <Route path="*">
                {/* Check we've got a logged in user otherwise redirect back to signin */}
                <AuthCheck>
                  <AppContextProvider
                    renderFooter
                    applicationsClient={appsClient}
                  >
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

                              return (
                                <ApplicationDetail
                                  name={params.name as string}
                                />
                              );
                            }}
                          />
                          <Route
                            exact
                            path={PageRoute.ApplicationAdd}
                            component={ApplicationAdd}
                          />
                          <Route
                            exact
                            path={PageRoute.GitlabOAuthCallback}
                            component={({ location }) => {
                              const params = qs.parse(location.search);

                              return (
                                <OAuthCallback
                                  provider={GitProvider.GitLab}
                                  code={params.code as string}
                                />
                              );
                            }}
                          />
                          <Route
                            exact
                            path={PageRoute.ApplicationRemove}
                            component={({ location }) => {
                              const params = qs.parse(location.search);

                              return (
                                <ApplicationRemove
                                  name={params.name as string}
                                />
                              );
                            }}
                          />
                          <Redirect
                            exact
                            from="/"
                            to={PageRoute.Applications}
                          />
                          <Route exact path="*" component={Error} />
                        </Switch>
                      </ErrorBoundary>
                      <ToastContainer
                        position="top-center"
                        autoClose={5000}
                        newestOnTop={false}
                      />
                    </Layout>
                  </AppContextProvider>
                </AuthCheck>
              </Route>
            </Switch>
          </AuthContextProvider>
        </Router>
      </ThemeProvider>
    </MuiThemeProvider>
  );
}
