import { SourceRefSourceKind } from "./lib/api/core/types.pb";
import Button from "./components/Button";
import Footer from "./components/Footer";
import GithubDeviceAuthModal from "./components/GithubDeviceAuthModal";
import Icon, { IconType } from "./components/Icon";
import LoadingPage from "./components/LoadingPage";
import RepoInputWithAuth from "./components/RepoInputWithAuth";
import UserSettings from "./components/UserSettings";
import AutomationsTable from "./components/AutomationsTable";
import SourceDetail from "./components/SourceDetail";
import SourcesTable from "./components/SourcesTable";
import KustomizationDetail from "./components/KustomizationDetail";
import HelmReleaseDetail from "./components/HelmReleaseDetail";
import Interval from "./components/Interval";
import Timestamp from "./components/Timestamp";
import AppContextProvider from "./contexts/AppContext";
import AuthContextProvider, { Auth, AuthCheck } from "./contexts/AuthContext";
import CallbackStateContextProvider from "./contexts/CallbackStateContext";
import {
  Automation,
  useListAutomations,
  useGetKustomization,
  useGetHelmRelease,
} from "./hooks/automations";
import { useIsAuthenticated } from "./hooks/gitprovider";
import { useListSources } from "./hooks/sources";
import { useFeatureFlags } from "./hooks/featureflags";
import FeatureFlagsContextProvider, {
  FeatureFlags,
} from "./contexts/FeatureFlags";
import { Applications as applicationsClient } from "./lib/api/applications/applications.pb";
import { Core as coreClient } from "./lib/api/core/core.pb";
import {
  clearCallbackState,
  getCallbackState,
  getProviderToken,
} from "./lib/storage";
import { muiTheme, theme } from "./lib/theme";
import { V2Routes } from "./lib/types";
import OAuthCallback from "./pages/OAuthCallback";
import SignIn from "./pages/SignIn";

export {
  AppContextProvider,
  applicationsClient,
  Auth,
  AuthContextProvider,
  AuthCheck,
  Automation,
  AutomationsTable,
  Button,
  CallbackStateContextProvider,
  clearCallbackState,
  coreClient,
  FeatureFlagsContextProvider,
  FeatureFlags,
  Footer,
  getCallbackState,
  getProviderToken,
  GithubDeviceAuthModal,
  HelmReleaseDetail,
  Icon,
  IconType,
  Interval,
  KustomizationDetail,
  LoadingPage,
  muiTheme,
  OAuthCallback,
  RepoInputWithAuth,
  SignIn,
  SourceDetail,
  SourceRefSourceKind,
  SourcesTable,
  theme,
  Timestamp,
  useIsAuthenticated,
  useListSources,
  useFeatureFlags,
  useGetKustomization,
  useGetHelmRelease,
  useListAutomations,
  UserSettings,
  V2Routes,
};
