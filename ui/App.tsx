import { MuiThemeProvider } from "@material-ui/core";
import qs from "query-string";
import * as React from "react";
import { QueryClient, QueryClientProvider } from "react-query";
import {
  BrowserRouter as Router,
  Redirect,
  // Route,
  Switch,
} from "react-router-dom";
import { CompatRouter, CompatRoute } from "react-router-dom-v5-compat";
import { ToastContainer } from "react-toastify";
import "react-toastify/dist/ReactToastify.css";
import { ThemeProvider } from "styled-components";
import ErrorBoundary from "./components/ErrorBoundary";
import Layout from "./components/Layout";
import Pendo from "./components/Pendo";
import AppContextProvider from "./contexts/AppContext";
import AuthContextProvider, { AuthCheck } from "./contexts/AuthContext";
import CoreClientContextProvider from "./contexts/CoreClientContext";
import { Core } from "./lib/api/core/core.pb";
import Fonts from "./lib/fonts";
import theme, { GlobalStyle, muiTheme } from "./lib/theme";
import { V2Routes } from "./lib/types";
import Error from "./pages/Error";
import SignIn from "./pages/SignIn";
import Automations from "./pages/v2/Automations";
import BucketDetail from "./pages/v2/BucketDetail";
import FluxRuntime from "./pages/v2/FluxRuntime";
import GitRepositoryDetail from "./pages/v2/GitRepositoryDetail";
import HelmChartDetail from "./pages/v2/HelmChartDetail";
import HelmReleasePage from "./pages/v2/HelmReleasePage";
import HelmRepositoryDetail from "./pages/v2/HelmRepositoryDetail";
import KustomizationPage from "./pages/v2/KustomizationPage";
import Notifications from "./pages/v2/Notifications";
import OCIRepositoryPage from "./pages/v2/OCIRepositoryPage";
import ProviderPage from "./pages/v2/ProviderPage";
import Sources from "./pages/v2/Sources";

const queryClient = new QueryClient();

function withSearchParams(Cmp) {
  return ({ location: { search }, ...rest }) => {
    const params = qs.parse(search);

    return <Cmp {...rest} {...params} />;
  };
}

const App = () => (
  <Layout>
    <ErrorBoundary>
      <Switch>
        <CompatRoute
          exact
          path={V2Routes.Automations}
          component={Automations}
        />
        <CompatRoute
          path={V2Routes.Kustomization}
          component={withSearchParams(KustomizationPage)}
        />
        <CompatRoute path={V2Routes.Sources} component={Sources} />
        <CompatRoute path={V2Routes.FluxRuntime} component={FluxRuntime} />
        <CompatRoute
          path={V2Routes.GitRepo}
          component={withSearchParams(GitRepositoryDetail)}
        />
        <CompatRoute
          path={V2Routes.HelmRepo}
          component={withSearchParams(HelmRepositoryDetail)}
        />
        <CompatRoute
          path={V2Routes.Bucket}
          component={withSearchParams(BucketDetail)}
        />
        <CompatRoute
          path={V2Routes.HelmRelease}
          component={withSearchParams(HelmReleasePage)}
        />
        <CompatRoute
          path={V2Routes.HelmChart}
          component={withSearchParams(HelmChartDetail)}
        />
        <CompatRoute
          path={V2Routes.OCIRepository}
          component={withSearchParams(OCIRepositoryPage)}
        />
        <CompatRoute
          path={V2Routes.Notifications}
          component={withSearchParams(Notifications)}
        />
        <CompatRoute
          path={V2Routes.Provider}
          component={withSearchParams(ProviderPage)}
        />
        <Redirect exact from="/" to={V2Routes.Automations} />
        <CompatRoute exact path="*" component={Error} />
      </Switch>
    </ErrorBoundary>
    <ToastContainer
      position="top-center"
      autoClose={5000}
      newestOnTop={false}
    />
  </Layout>
);

export default function AppContainer() {
  return (
    <MuiThemeProvider theme={muiTheme}>
      <QueryClientProvider client={queryClient}>
        <ThemeProvider theme={theme}>
          <Fonts />
          <GlobalStyle />
          <Router>
            <CompatRouter>
              <AppContextProvider renderFooter>
                <AuthContextProvider>
                  <CoreClientContextProvider api={Core}>
                    <Pendo defaultTelemetryFlag="false" />
                    <Switch>
                      {/* <Signin> does not use the base page <Layout> so pull it up here */}
                      <CompatRoute component={SignIn} exact path="/sign_in" />
                      <CompatRoute path="" component={SignIn}>
                        {/* Check we've got a logged in user otherwise redirect back to signin */}
                        <AuthCheck>
                          <App />
                        </AuthCheck>
                      </CompatRoute>
                    </Switch>
                  </CoreClientContextProvider>
                </AuthContextProvider>
              </AppContextProvider>
            </CompatRouter>
          </Router>
        </ThemeProvider>
      </QueryClientProvider>
    </MuiThemeProvider>
  );
}
