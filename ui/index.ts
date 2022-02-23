import Button from "./components/Button";
import Footer from "./components/Footer";
import GithubDeviceAuthModal from "./components/GithubDeviceAuthModal";
import LoadingPage from "./components/LoadingPage";
import RepoInputWithAuth from "./components/RepoInputWithAuth";
import Icon, { IconType } from "./components/Icon";
import AppContextProvider from "./contexts/AppContext";
import AuthContextProvider from "./contexts/AuthContext";
import CallbackStateContextProvider from "./contexts/CallbackStateContext";
import useApplications from "./hooks/applications";
import { Applications as applicationsClient } from "./lib/api/applications/applications.pb";
import {
  clearCallbackState,
  getCallbackState,
  getProviderToken,
} from "./lib/storage";
import { theme, muiTheme } from "./lib/theme";
import ApplicationDetail from "./pages/ApplicationDetail";
import Applications from "./pages/Applications";
import OAuthCallback from "./pages/OAuthCallback";

export {
  AuthContextProvider,
  AppContextProvider,
  ApplicationDetail,
  Applications,
  applicationsClient,
  getProviderToken,
  GithubDeviceAuthModal,
  LoadingPage,
  theme,
  muiTheme,
  useApplications,
  Footer,
  RepoInputWithAuth,
  CallbackStateContextProvider,
  getCallbackState,
  clearCallbackState,
  OAuthCallback,
  Button,
  Icon,
  IconType,
};
