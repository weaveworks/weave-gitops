import { MuiThemeProvider } from "@material-ui/core";
import qs from "query-string";
import * as React from "react";
import { QueryClient, QueryClientProvider } from "react-query";
import {
  Routes,
  Route,
  BrowserRouter as Router,
  useLocation,
  Navigate,
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
import ProfileInfo from "./components/ProfileSettings";
import AppContextProvider from "./contexts/AppContext";
import AuthContextProvider, { AuthCheck } from "./contexts/AuthContext";
import CoreClientContextProvider from "./contexts/CoreClientContext";
import { Core } from "./lib/api/core/core.pb";
import Fonts from "./lib/fonts";
import { Redirect } from "./lib/nav";
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

function withSearchParams() {
  const location = useLocation();
  const params = qs.parse(location.search);
  return params;
}

const App = () => (
  <Layout>
    <PendoContainer />
    <ErrorBoundary>
      <Routes>
        <Route path={V2Routes.Automations} element={<Automations />} />
        <Route
          path={V2Routes.Kustomization + "/*"}
          element={<KustomizationPage {...withSearchParams()} />}
        />
        <Route path={V2Routes.Sources} element={<Sources />} />
        <Route
          path={V2Routes.ImageAutomation + "/*"}
          element={<ImageAutomationPage />}
        />
        <Route
          path={V2Routes.ImageAutomationUpdatesDetails + "/*"}
          element={<ImageAutomationUpdatesDetails {...withSearchParams()} />}
        />
        <Route
          path={V2Routes.ImageAutomationRepositoryDetails + "/*"}
          element={<ImageAutomationRepoDetails {...withSearchParams()} />}
        />
        <Route
          path={V2Routes.ImagePolicyDetails + "/*"}
          element={<ImagePolicyDetails {...withSearchParams()} />}
        />
        <Route path={V2Routes.FluxRuntime + "/*"} element={<FluxRuntime />} />
        <Route
          path={V2Routes.GitRepo + "/*"}
          element={<GitRepositoryDetail {...withSearchParams()} />}
        />
        <Route
          path={V2Routes.HelmRepo + "/*"}
          element={<HelmRepositoryDetail {...withSearchParams()} />}
        />
        <Route
          path={V2Routes.Bucket + "/*"}
          element={<BucketDetail {...withSearchParams()} />}
        />
        <Route
          path={V2Routes.HelmRelease + "/*"}
          element={<HelmReleasePage {...withSearchParams()} />}
        />
        <Route
          path={V2Routes.HelmChart + "/*"}
          element={<HelmChartDetail {...withSearchParams()} />}
        />
        <Route
          path={V2Routes.OCIRepository + "/*"}
          element={<OCIRepositoryPage {...withSearchParams()} />}
        />
        <Route
          path={V2Routes.Notifications + "/*"}
          element={<Notifications {...withSearchParams()} />}
        />
        <Route
          path={V2Routes.Provider + "/*"}
          element={<ProviderPage {...withSearchParams()} />}
        />
        <Route path={V2Routes.UserInfo} element={UserInfo} />

        <Route path="/" element={<Navigate to="/applications" replace />} />

        <Route path="*" element={<Error />} />
      </Routes>
    </ErrorBoundary>
    <ToastContainer
      position="top-center"
      autoClose={5000}
      newestOnTop={false}
    />
  </Layout>
);

function AuthCheckApp() {
  return (
    <AuthCheck>
      <App />
    </AuthCheck>
  );
}

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
                  <Routes>
                    {/* <Signin> does not use the base page <Layout> so pull it up here */}
                    <Route path="/sign_in" element={<SignIn />} />
                    {/* path="" matches all the paths */}
                    <Route path="*" element={<AuthCheckApp />} />
                  </Routes>
                </CoreClientContextProvider>
              </AuthContextProvider>
            </AppContextProvider>
          </Router>
        </ThemeProvider>
      </QueryClientProvider>
    </MuiThemeProvider>
  );
}
