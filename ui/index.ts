import Footer from "./components/Footer";
import GithubDeviceAuthModal from "./components/GithubDeviceAuthModal";
import LoadingPage from "./components/LoadingPage";
import AppContextProvider from "./contexts/AppContext";
import useApplications from "./hooks/applications";
import { Applications as applicationsClient } from "./lib/api/applications/applications.pb";
import { getProviderToken } from "./lib/storage";
import theme from "./lib/theme";
import ApplicationAdd from "./pages/ApplicationAdd";
import ApplicationDetail from "./pages/ApplicationDetail";
import Applications from "./pages/Applications";

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
  Footer,
};
