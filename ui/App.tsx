import { MuiThemeProvider } from "@material-ui/core";
import qs from "query-string";
import * as React from "react";
import { QueryClient, QueryClientProvider } from "react-query";
import {
  Redirect,
  Route,
  BrowserRouter as Router,
  Switch,
} from "react-router-dom";
import { ToastContainer } from "react-toastify";
import "react-toastify/dist/ReactToastify.css";
import { ThemeProvider } from "styled-components";
import ErrorBoundary from "./components/ErrorBoundary";
import ImagePolicyDetails from "./components/ImageAutomation/policies/ImagePolicyDetails";
import ImageAutomationRepoDetails from "./components/ImageAutomation/repositories/ImageAutomationRepoDetails";
import ImageAutomationUpdatesDetails from "./components/ImageAutomation/updates/ImageAutomationUpdatesDetails";
import Layout from "./components/Layout";
import PendoContainer from "./components/PendoContainer";
import AppContextProvider, {
  AppContext,
  ThemeTypes,
} from "./contexts/AppContext";
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
import ImageAutomationPage from "./pages/v2/ImageAutomationPage";
import KustomizationPage from "./pages/v2/KustomizationPage";
import Notifications from "./pages/v2/Notifications";
import OCIRepositoryPage from "./pages/v2/OCIRepositoryPage";
import ProviderPage from "./pages/v2/ProviderPage";
import Sources from "./pages/v2/Sources";
import UserInfo from "./pages/v2/UserInfo";

const queryClient = new QueryClient();

function withSearchParams(Cmp) {
  return ({ location: { search }, ...rest }) => {
    const params = qs.parse(search);

    return <Cmp {...rest} {...params} />;
  };
}

const App = () => {
  const { settings } = React.useContext(AppContext);
  const dark = settings.theme === ThemeTypes.Dark;
  return (
    <Layout>
      <PendoContainer />
      <ErrorBoundary>
        <Switch>
          <Route exact path={V2Routes.Automations} component={Automations} />
          <Route
            path={V2Routes.Kustomization}
            component={withSearchParams(KustomizationPage)}
          />
          <Route path={V2Routes.Sources} component={Sources} />
          <Route
            path={V2Routes.ImageAutomation}
            component={ImageAutomationPage}
          />
          <Route
            path={V2Routes.ImageAutomationUpdatesDetails}
            component={withSearchParams(ImageAutomationUpdatesDetails)}
          />
          <Route
            path={V2Routes.ImageAutomationRepositoryDetails}
            component={withSearchParams(ImageAutomationRepoDetails)}
          />
          <Route
            path={V2Routes.ImagePolicyDetails}
            component={withSearchParams(ImagePolicyDetails)}
          />
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
          <Route
            path={V2Routes.OCIRepository}
            component={withSearchParams(OCIRepositoryPage)}
          />
          <Route
            path={V2Routes.Notifications}
            component={withSearchParams(Notifications)}
          />
          <Route
            path={V2Routes.Provider}
            component={withSearchParams(ProviderPage)}
          />
          <Route path={V2Routes.UserInfo} component={UserInfo} />

          <Redirect exact from="/" to={V2Routes.Automations} />

          <Route exact path="*" component={Error} />
        </Switch>
      </ErrorBoundary>
      <ToastContainer
        position="top-center"
        autoClose={5000}
        newestOnTop={false}
        theme={dark ? ThemeTypes.Dark : ThemeTypes.Light}
      />
    </Layout>
  );
};

const StylesProvider = ({ children }) => {
  const { settings } = React.useContext(AppContext);
  const mode = settings.theme;
  const appliedTheme = theme(mode);
  return (
    <ThemeProvider theme={appliedTheme}>
      <MuiThemeProvider theme={muiTheme(appliedTheme.colors)}>
        <Fonts />
        <GlobalStyle />
        {children}
      </MuiThemeProvider>
    </ThemeProvider>
  );
};

export default function AppContainer() {
  return (
    <QueryClientProvider client={queryClient}>
      <Router>
        <AppContextProvider renderFooter>
          <StylesProvider>
            <AuthContextProvider>
              <CoreClientContextProvider api={Core}>
                <Switch>
                  {/* <Signin> does not use the base page <Layout> so pull it up here */}
                  <Route exact path="/sign_in">
                    <SignIn />
                  </Route>
                  <Route path="*">
                    {/* Check we've got a logged in user otherwise redirect back to signin */}
                    <AuthCheck>
                      <App />
                    </AuthCheck>
                  </Route>
                </Switch>
              </CoreClientContextProvider>
            </AuthContextProvider>
          </StylesProvider>
        </AppContextProvider>
      </Router>
    </QueryClientProvider>
  );
}
