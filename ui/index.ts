import _AppContextProvider from "./contexts/AppContext";
import { Applications as appsClient } from "./lib/api/applications/applications.pb";
import _Theme from "./lib/theme";
import _ApplicationDetail from "./pages/ApplicationDetail";
import _Applications from "./pages/Applications";

export const theme = _Theme;
export const AppContextProvider = _AppContextProvider;
export const Applications = _Applications;
export const ApplicationDetail = _ApplicationDetail;
export const applicationsClient = appsClient;
