import {
  ThemeProvider as MuiThemeProvider,
  StyledEngineProvider,
} from "@mui/material";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import qs from "query-string";
import * as React from "react";
import {
  Routes,
  Route,
  BrowserRouter,
  useLocation,
  Navigate,
} from "react-router";
import { ToastContainer } from "react-toastify";
import "react-toastify/dist/ReactToastify.css";
import { ThemeProvider } from "styled-components";
import ErrorBoundary from "./components/ErrorBoundary";
import Footer from "./components/Footer";
import { IconType } from "./components/Icon";
import ImagePolicyDetails from "./components/ImageAutomation/policies/ImagePolicyDetails";
import ImageAutomationRepoDetails from "./components/ImageAutomation/repositories/ImageAutomationRepoDetails";
import ImageAutomationUpdatesDetails from "./components/ImageAutomation/updates/ImageAutomationUpdatesDetails";
import Layout from "./components/Layout";
import Logo from "./components/Logo";
import Nav, { NavItem } from "./components/Nav";
import PolicyViolationPage from "./components/Policies/PolicyViolations/PolicyViolationPage";
import AppContextProvider, {
  AppContext,
  ThemeTypes,
} from "./contexts/AppContext";
import AuthContextProvider, { AuthCheck } from "./contexts/AuthContext";
import CoreClientContextProvider from "./contexts/CoreClientContext";
import {
  LinkResolverProvider,
  useLinkResolver,
} from "./contexts/LinkResolverContext";
import { useFeatureFlags } from "./hooks/featureflags";
import useNavigation from "./hooks/navigation";
import { useInDarkMode } from "./hooks/theme";
import { Core } from "./lib/api/core/core.pb";
import Fonts from "./lib/fonts";
import { getParentNavRouteValue } from "./lib/nav";
import theme, { GlobalStyle, muiTheme } from "./lib/theme";
import { V2Routes } from "./lib/types";
import { getBasePath } from "./lib/utils";
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
import PoliciesList from "./pages/v2/PoliciesList";
import PolicyDetailsPage from "./pages/v2/PolicyDetailsPage";
import ProviderPage from "./pages/v2/ProviderPage";
import Runtime from "./pages/v2/Runtime";
import Sources from "./pages/v2/Sources";
import UserInfo from "./pages/v2/UserInfo";

const queryClient = new QueryClient();

const WithSearchParams = ({
  component: Component,
  ...props
}: {
  component: React.FunctionComponent<any>;
}) => {
  const location = useLocation();
  const params = qs.parse(location.search);

  return <Component {...props} {...params} />;
};

function getRuntimeNavItem(isNewRuntimeEnabled: boolean): NavItem {
  if (isNewRuntimeEnabled) {
    return {
      label: "Runtime",
      link: { value: V2Routes.Runtime },
      icon: IconType.FluxIcon,
    };
  }

  return {
    label: "Flux Runtime",
    link: { value: V2Routes.FluxRuntime },
    icon: IconType.FluxIcon,
  };
}

const App = () => {
  const dark = useInDarkMode();

  const { isFlagEnabled } = useFeatureFlags();

  const isNewRuntimeEnabled = isFlagEnabled(
    "WEAVE_GITOPS_FEATURE_GITOPS_RUNTIME",
  );

  const navItems: NavItem[] = [
    {
      label: "Applications",
      link: { value: V2Routes.Automations },
      icon: IconType.ApplicationsIcon,
    },
    {
      label: "Sources",
      link: { value: V2Routes.Sources },
      icon: IconType.SourcesIcon,
    },
    {
      label: "Image Automation",
      link: { value: V2Routes.ImageAutomation },
      icon: IconType.ImageAutomationIcon,
    },
    {
      label: "Policies",
      link: { value: V2Routes.Policies },
      icon: IconType.PoliciesIcon,
    },
    getRuntimeNavItem(isNewRuntimeEnabled),
    {
      label: "Notifications",
      link: { value: V2Routes.Notifications },
      icon: IconType.NotificationsIcon,
    },
  ];

  const [collapsed, setCollapsed] = React.useState<boolean>(false);
  const { currentPage } = useNavigation();
  const value = getParentNavRouteValue(currentPage);

  const logo = <Logo collapsed={collapsed} link={V2Routes.Automations} />;

  const nav = (
    <Nav
      navItems={navItems}
      collapsed={collapsed}
      setCollapsed={setCollapsed}
      currentPage={value}
    />
  );

  return (
    <Layout logo={logo} nav={nav}>
      <ErrorBoundary>
        <Routes>
          <Route path={V2Routes.Automations} element={<Automations />} />
          <Route path={V2Routes.Kustomization}>
            <Route
              element={<WithSearchParams component={KustomizationPage} />}
              index
              path="*"
            />
          </Route>

          <Route path={V2Routes.Sources}>
            <Route
              element={<WithSearchParams component={Sources} />}
              index
              path="*"
            />
          </Route>

          <Route path={V2Routes.ImageAutomation}>
            <Route
              element={<WithSearchParams component={ImageAutomationPage} />}
              index
              path="*"
            />
          </Route>

          <Route path={V2Routes.ImageAutomationUpdatesDetails}>
            <Route
              element={
                <WithSearchParams component={ImageAutomationUpdatesDetails} />
              }
              index
              path="*"
            />
          </Route>

          <Route path={V2Routes.ImageAutomationRepositoryDetails}>
            <Route
              element={
                <WithSearchParams component={ImageAutomationRepoDetails} />
              }
              index
              path="*"
            />
          </Route>

          <Route path={V2Routes.ImagePolicyDetails}>
            <Route
              element={<WithSearchParams component={ImagePolicyDetails} />}
              index
              path="*"
            />
          </Route>

          {isNewRuntimeEnabled ? (
            <Route path={V2Routes.Runtime}>
              <Route index path="*" element={<Runtime />} />
            </Route>
          ) : (
            <Route path={V2Routes.FluxRuntime}>
              <Route index path="*" element={<FluxRuntime />} />
            </Route>
          )}

          <Route path={V2Routes.GitRepo}>
            <Route
              element={<WithSearchParams component={GitRepositoryDetail} />}
              index
              path="*"
            />
          </Route>
          <Route path={V2Routes.HelmRepo}>
            <Route
              element={<WithSearchParams component={HelmRepositoryDetail} />}
              index
              path="*"
            />
          </Route>
          <Route path={V2Routes.Bucket}>
            <Route
              element={<WithSearchParams component={BucketDetail} />}
              index
              path="*"
            />
          </Route>
          <Route path={V2Routes.HelmRelease}>
            <Route
              element={<WithSearchParams component={HelmReleasePage} />}
              index
              path="*"
            />
          </Route>
          <Route path={V2Routes.HelmChart}>
            <Route
              element={<WithSearchParams component={HelmChartDetail} />}
              index
              path="*"
            />
          </Route>
          <Route path={V2Routes.OCIRepository}>
            <Route
              element={<WithSearchParams component={OCIRepositoryPage} />}
              index
              path="*"
            />
          </Route>
          <Route path={V2Routes.Notifications}>
            <Route
              element={<WithSearchParams component={Notifications} />}
              index
              path="*"
            />
          </Route>
          <Route path={V2Routes.Provider}>
            <Route
              element={<WithSearchParams component={ProviderPage} />}
              index
              path="*"
            />
          </Route>
          <Route path={V2Routes.PolicyViolationDetails}>
            <Route
              element={<WithSearchParams component={PolicyViolationPage} />}
              index
              path="*"
            />
          </Route>

          <Route path={V2Routes.UserInfo} element={<UserInfo />} />
          <Route path={V2Routes.Policies} element={<PoliciesList />} />
          <Route
            element={<WithSearchParams component={PolicyDetailsPage} />}
            path={V2Routes.PolicyDetailsPage}
          />
          <Route
            element={<Navigate to={V2Routes.Automations} replace />}
            path="/"
          />
          <Route path="*" element={Error} />
        </Routes>
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
      <StyledEngineProvider injectFirst>
        <MuiThemeProvider theme={muiTheme(appliedTheme.colors, mode)}>
          <Fonts />
          <GlobalStyle />
          {children}
        </MuiThemeProvider>
      </StyledEngineProvider>
    </ThemeProvider>
  );
};

export default function AppContainer() {
  const resolver = useLinkResolver();
  return (
    <QueryClientProvider client={queryClient}>
      <BrowserRouter basename={getBasePath()}>
        <AppContextProvider footer={<Footer />}>
          <StylesProvider>
            <AuthContextProvider>
              <CoreClientContextProvider api={Core}>
                <LinkResolverProvider resolver={resolver}>
                  <Routes>
                    <Route element={<SignIn />} path="/sign_in" />
                    {/* Check we've got a logged in user otherwise redirect back to signin */}
                    <Route
                      path="*"
                      element={
                        <AuthCheck>
                          <App />
                        </AuthCheck>
                      }
                    />
                  </Routes>
                  <ToastContainer
                    position="top-center"
                    autoClose={5000}
                    newestOnTop={false}
                  />
                </LinkResolverProvider>
              </CoreClientContextProvider>
            </AuthContextProvider>
          </StylesProvider>
        </AppContextProvider>
      </BrowserRouter>
    </QueryClientProvider>
  );
}
