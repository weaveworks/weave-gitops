import Button from "./components/Button";
import Footer from "./components/Footer";
import GithubDeviceAuthModal from "./components/GithubDeviceAuthModal";
import Icon, { IconType } from "./components/Icon";
import LoadingPage from "./components/LoadingPage";
import RepoInputWithAuth from "./components/RepoInputWithAuth";
import UserSettings from "./components/UserSettings";
import AppContextProvider from "./contexts/AppContext";
<<<<<<< HEAD
import AuthContextProvider, { Auth, AuthCheck } from "./contexts/AuthContext";
=======
import AuthContextProvider from "./contexts/AuthContext";
>>>>>>> b4662b16 (Fix linting errors)
import CallbackStateContextProvider from "./contexts/CallbackStateContext";
import FeatureFlagsContextProvider, {
  FeatureFlags,
} from "./contexts/FeatureFlags";
import { Applications as applicationsClient } from "./lib/api/applications/applications.pb";
import {
  clearCallbackState,
  getCallbackState,
  getProviderToken,
} from "./lib/storage";
import { muiTheme, theme } from "./lib/theme";
import OAuthCallback from "./pages/OAuthCallback";
import SignIn from "./pages/SignIn";

export {
  FeatureFlagsContextProvider,
  FeatureFlags,
  AuthContextProvider,
  AuthCheck,
  AppContextProvider,
  applicationsClient,
  getProviderToken,
  GithubDeviceAuthModal,
  LoadingPage,
  theme,
  muiTheme,
  Footer,
  RepoInputWithAuth,
  CallbackStateContextProvider,
  getCallbackState,
  clearCallbackState,
  OAuthCallback,
  Button,
  Icon,
  IconType,
  UserSettings,
  SignIn,
  Auth,
};
