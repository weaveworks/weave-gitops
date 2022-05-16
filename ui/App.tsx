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
import AuthContextProvider, { AuthCheck } from "./contexts/AuthContext";
import CoreClientContextProvider from "./contexts/CoreClientContext";
import VersionContextProvider from "./contexts/VersionContext";
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
        <Route exact path={V2Routes.Automations} component={Automations} />
        <Route
          path={V2Routes.Kustomization}
          component={withSearchParams(KustomizationPage)}
        />
        <Route path={V2Routes.Sources} component={Sources} />
        <Route path={V2Routes.FluxRuntime} component={FluxRuntime} />
        <Route
          path={V2Routes.GitRepo}
          component={withSearchParams(GitRepositoryDetail)}
        />
        <Route
          path={V2Routes.HelmRepo}
          component={withSearchParams(HelmRepositoryDetail)}
        />
        <Route
          path={V2Routes.Bucket}
          component={withSearchParams(BucketDetail)}
        />
        <Route
          path={V2Routes.HelmRelease}
          component={withSearchParams(HelmReleasePage)}
        />
        <Route
          path={V2Routes.HelmChart}
          component={withSearchParams(HelmChartDetail)}
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
);

export default function AppContainer() {
  return (
    <MuiThemeProvider theme={muiTheme}>
      <QueryClientProvider client={queryClient}>
        <ThemeProvider theme={theme}>
          <Fonts />
          <GlobalStyle />
          <Router>
            <AppContextProvider renderFooter>
              <AuthContextProvider>
                <CoreClientContextProvider api={Core}>
                  <VersionContextProvider>
                    <Switch>
                      {/* <Signin> does not use the base page <Layout> so pull it up here */}
                      <Route component={SignIn} exact path="/sign_in" />
                      <Route path="*">
                        {/* Check we've got a logged in user otherwise redirect back to signin */}
                        <AuthCheck>
                          <App />
                        </AuthCheck>
                      </Route>
                    </Switch>
                  </VersionContextProvider>
                </CoreClientContextProvider>
              </AuthContextProvider>
            </AppContextProvider>
          </Router>
        </ThemeProvider>
      </QueryClientProvider>
    </MuiThemeProvider>
  );
}
