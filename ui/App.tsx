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
import FeatureFlagsContextProvider from "./contexts/FeatureFlags";
import { Core } from "./lib/api/core/core.pb";
import Fonts from "./lib/fonts";
import theme, { GlobalStyle, muiTheme } from "./lib/theme";
import { V2Routes } from "./lib/types";
import SignIn from "./pages/SignIn";
import Automations from "./pages/v2/Automations";
import FluxRuntime from "./pages/v2/FluxRuntime";
import GitRepositoryDetail from "./pages/v2/GitRepositoryDetail";
import KustomizationDetail from "./pages/v2/KustomizationDetail";
import Sources from "./pages/v2/Sources";

const queryClient = new QueryClient();

function withName(Cmp) {
  return ({ location: { search }, ...rest }) => {
    const params = qs.parse(search);

    return <Cmp {...rest} name={params.name as string} />;
  };
}

const App = () => (
  <Layout>
    <ErrorBoundary>
      <Switch>
        <Route exact path={V2Routes.Automations} component={Automations} />
        <Route
          exact
          path={V2Routes.Kustomization}
          component={withName(KustomizationDetail)}
        />
        <Route exact path={V2Routes.Sources} component={Sources} />
        <Route exact path={V2Routes.FluxRuntime} component={FluxRuntime} />
        <Route
          exact
          path={V2Routes.GitRepo}
          component={withName(GitRepositoryDetail)}
        />
        <Redirect exact from="/" to={V2Routes.Automations} />
        {/* <Route exact path="*" component={Error} /> */}
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
      <AppContextProvider renderFooter coreClient={Core}>
        <QueryClientProvider client={queryClient}>
          <ThemeProvider theme={theme}>
            <Fonts />
            <GlobalStyle />
            <Router>
              <FeatureFlagsContextProvider>
                <AuthContextProvider>
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
                </AuthContextProvider>
              </FeatureFlagsContextProvider>
            </Router>
          </ThemeProvider>
        </QueryClientProvider>
      </AppContextProvider>
    </MuiThemeProvider>
  );
}
