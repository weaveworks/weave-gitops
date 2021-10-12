import _LoadingPage from "./components/LoadingPage";
import _AppContextProvider from "./contexts/AppContext";
import _useApplications from "./hooks/applications";
import { Applications as appsClient } from "./lib/api/applications/applications.pb";
import _Theme from "./lib/theme";
import _ApplicationDetail from "./pages/ApplicationDetail";
import _ApplicationAdd from "./pages/ApplicationAdd";
import _Applications from "./pages/Applications";
import _GithubDeviceAuthModal from "./components/GithubDeviceAuthModal";
import { getProviderToken as _getProviderToken } from "./lib/utils";

export const theme = _Theme;
export const AppContextProvider = _AppContextProvider;
export const Applications = _Applications;
export const ApplicationDetail = _ApplicationDetail;
export const ApplicationAdd = _ApplicationAdd;
export const applicationsClient = appsClient;
export const useApplications = _useApplications;
export const LoadingPage = _LoadingPage;

// Authy things
export const GithubDeviceAuthModal = _GithubDeviceAuthModal;
export const getProviderToken = _getProviderToken;
