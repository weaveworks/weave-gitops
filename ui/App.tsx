import { MuiThemeProvider } from "@material-ui/core";
import qs from "query-string";
import * as React from "react";
import { QueryClient, QueryClientProvider } from "react-query";
import {
  BrowserRouter as Router,
  Redirect,
  Route,
  Switch
} from "react-router-dom";
import { ToastContainer } from "react-toastify";
import "react-toastify/dist/ReactToastify.css";
import { ThemeProvider } from "styled-components";
import ErrorBoundary from "./components/ErrorBoundary";
import Layout from "./components/Layout";
import AppContextProvider from "./contexts/AppContext";
import { Core } from "./lib/api/core/core.pb";
import Fonts from "./lib/fonts";
import theme, { GlobalStyle, muiTheme } from "./lib/theme";
import { V2Routes } from "./lib/types";
import Error from "./pages/Error";
import Automations from "./pages/v2/Automations";
import FluxRuntime from "./pages/v2/FluxRuntime";
import KustomizationDetail from "./pages/v2/KustomizationDetail";
import Sources from "./pages/v2/Sources";

const queryClient = new QueryClient();

function withName(Cmp) {
  return ({ location: { search }, ...rest }) => {
    const params = qs.parse(search);

    return <Cmp {...rest} name={params.name as string} />;
  };
}

export default function App() {
  return (
    <MuiThemeProvider theme={muiTheme}>
      <ThemeProvider theme={theme}>
        <QueryClientProvider client={queryClient}>
          <Fonts />
          <GlobalStyle />
          <Router>
            <AppContextProvider renderFooter coreClient={Core}>
              <Layout>
                <ErrorBoundary>
                  <Switch>
                    <Route
                      exact
<<<<<<< HEAD
                      path={V2Routes.Automations}
                      component={Automations}
=======
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
>>>>>>> Delete GetApplication, ListApplications endpoints (#1455)
                    />
                    <Route
                      exact
                      path={V2Routes.Kustomization}
                      component={withName(KustomizationDetail)}
                    />
                    <Route exact path={V2Routes.Sources} component={Sources} />
                    <Route
                      exact
                      path={V2Routes.FluxRuntime}
                      component={FluxRuntime}
                    />
                    <Redirect exact from="/" to={V2Routes.Automations} />
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
