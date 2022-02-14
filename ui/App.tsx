import { MuiThemeProvider } from "@material-ui/core";
import qs from "query-string";
import * as React from "react";
import { QueryClient, QueryClientProvider } from "react-query";
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
import { Apps } from "./lib/api/app/apps.pb";
import {
  Applications as appsClient,
  GitProvider,
} from "./lib/api/applications/applications.pb";
import Fonts from "./lib/fonts";
import theme, { GlobalStyle, muiTheme } from "./lib/theme";
import { PageRoute, V2Routes } from "./lib/types";
import ApplicationDetail from "./pages/ApplicationDetail";
import Applications from "./pages/Applications";
import Error from "./pages/Error";
import OAuthCallback from "./pages/OAuthCallback";
import AddAutomation from "./pages/v2/AddAutomation";
import AddGitRepo from "./pages/v2/AddGitRepo";
import AddKustomization from "./pages/v2/AddKustomization";
import AddSource from "./pages/v2/AddSource";
import Application from "./pages/v2/Application/Application";
import ApplicationList from "./pages/v2/ApplicationList/ApplicationList";
import KustomizationList from "./pages/v2/KustomizationList";
import NewApp from "./pages/v2/NewApp";
import SourcesList from "./pages/v2/SourcesList";

const queryClient = new QueryClient();

const withAppName =
  (Cmp) =>
  ({ location, ...rest }) => {
    const params = qs.parse(location.search);

    return <Cmp appName={params.appName as string} {...rest} />;
  };

export default function App() {
  return (
    <MuiThemeProvider theme={muiTheme}>
      <ThemeProvider theme={theme}>
        <QueryClientProvider client={queryClient}>
          <Fonts />
          <GlobalStyle />
          <Router>
            <AppContextProvider
              renderFooter
              applicationsClient={appsClient}
              appsClient={Apps}
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
                          <ApplicationDetail name={params.name as string} />
                        );
                      }}
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
                      path={V2Routes.ApplicationList}
                      component={ApplicationList}
                    />
                    <Route exact path={V2Routes.NewApp} component={NewApp} />
                    <Route
                      exact
                      path={V2Routes.Application}
                      component={({ location }) => {
                        const params = qs.parse(location.search);

                        return (
                          <Application
                            name={params.name as string}
                            namespace={params.namespace as string}
                          />
                        );
                      }}
                    />
                    <Route
                      exact
                      path={V2Routes.AddKustomization}
                      component={withAppName(AddKustomization)}
                    />
                    <Route
                      exact
                      path={V2Routes.AddSource}
                      component={withAppName(AddSource)}
                    />
                    <Route
                      exact
                      path={V2Routes.AddGitRepo}
                      component={withAppName(AddGitRepo)}
                    />
                    <Route
                      exact
                      path={V2Routes.AddAutomation}
                      component={withAppName(AddAutomation)}
                    />
                    <Route
                      exact
                      path={V2Routes.KustomizationList}
                      component={KustomizationList}
                    />
                    <Route
                      exact
                      path={V2Routes.SourcesList}
                      component={SourcesList}
                    />
                    <Redirect exact from="/" to={V2Routes.ApplicationList} />
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
          </Router>
        </QueryClientProvider>
      </ThemeProvider>
    </MuiThemeProvider>
  );
}
