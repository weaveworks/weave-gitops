import { SourceRefSourceKind } from "./lib/api/core/types.pb";
import Button from "./components/Button";
import Footer from "./components/Footer";
import GithubDeviceAuthModal from "./components/GithubDeviceAuthModal";
import Icon, { IconType } from "./components/Icon";
import LoadingPage from "./components/LoadingPage";
import RepoInputWithAuth from "./components/RepoInputWithAuth";
import UserSettings from "./components/UserSettings";
import AutomationsTable from "./components/AutomationsTable";
import BucketDetail from "./components/BucketDetail";
import FluxRuntime from "./components/FluxRuntime";
import GitRepositoryDetail from "./components/GitRepositoryDetail";
import HelmChartDetail from "./components/HelmChartDetail";
import HelmReleaseDetail from "./components/HelmReleaseDetail";
import HelmRepositoryDetail from "./components/HelmRepositoryDetail";
import KustomizationDetail from "./components/KustomizationDetail";
import Page from "./components/Page";
import SourcesTable from "./components/SourcesTable";
import Interval from "./components/Interval";
import Timestamp from "./components/Timestamp";
import { Field, SortType } from "./components/DataTable";
import FilterableTable, {
  filterConfigForStatus,
  filterConfigForString,
} from "./components/FilterableTable";
import AppContextProvider from "./contexts/AppContext";
import CoreClientContextProvider from "./contexts/CoreClientContext";
import AuthContextProvider, { Auth, AuthCheck } from "./contexts/AuthContext";
import CallbackStateContextProvider from "./contexts/CallbackStateContext";
import {
  Automation,
  useListAutomations,
  useGetKustomization,
  useGetHelmRelease,
} from "./hooks/automations";
import { useListFluxRuntimeObjects } from "./hooks/flux";
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
import KubeStatusIndicator from "./components/KubeStatusIndicator";
import { statusSortHelper } from "./lib/utils";

export {
  AppContextProvider,
  applicationsClient,
  Auth,
  AuthContextProvider,
  AuthCheck,
  Automation,
  AutomationsTable,
  BucketDetail,
  Button,
  CallbackStateContextProvider,
  clearCallbackState,
  coreClient,
  CoreClientContextProvider,
  FeatureFlagsContextProvider,
  FeatureFlags,
  Field,
  FilterableTable,
  filterConfigForString,
  filterConfigForStatus,
  FluxRuntime,
  Footer,
  getCallbackState,
  getProviderToken,
  GithubDeviceAuthModal,
  GitRepositoryDetail,
  HelmChartDetail,
  HelmReleaseDetail,
  HelmRepositoryDetail,
  Icon,
  IconType,
  Interval,
  KubeStatusIndicator,
  KustomizationDetail,
  LoadingPage,
  muiTheme,
  OAuthCallback,
  Page,
  RepoInputWithAuth,
  SignIn,
  statusSortHelper,
  SortType,
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
  useListFluxRuntimeObjects,
  UserSettings,
  V2Routes,
};
