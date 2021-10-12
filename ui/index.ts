import { Applications as applicationsClient } from "./lib/api/applications/applications.pb";
import { getProviderToken } from "./lib/utils";
import AppContextProvider from "./contexts/AppContext";
import ApplicationAdd from "./pages/ApplicationAdd";
import ApplicationDetail from "./pages/ApplicationDetail";
import Applications from "./pages/Applications";
import GithubDeviceAuthModal from "./components/GithubDeviceAuthModal";
import LoadingPage from "./components/LoadingPage";
import theme from "./lib/theme";
import useApplications from "./hooks/applications";

export {
  AppContextProvider,
  ApplicationAdd,
  ApplicationDetail,
  Applications,
  applicationsClient,
  getProviderToken,
  GithubDeviceAuthModal,
  LoadingPage,
  theme,
  useApplications,
};
