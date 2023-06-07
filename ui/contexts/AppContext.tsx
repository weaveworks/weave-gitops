import * as React from "react";
import { useHistory } from "react-router-dom";
import { DetailViewProps } from "../components/DetailModal";
import { formatURL } from "../lib/nav";
import { PageRoute, V2Routes } from "../lib/types";
import { notifySuccess } from "../lib/utils";

type AppState = {
  error: null | { fatal: boolean; message: string; detail?: string };
  detailModal: DetailViewProps;
};

export enum ThemeTypes {
  Light = "light",
  Dark = "dark",
}

type AppSettings = {
  renderFooter: boolean;
  theme: ThemeTypes;
};

export type AppContextType = {
  userConfigRepoName: string;
  doAsyncError: (message: string, detail: string) => void;
  clearAsyncError: () => void;
  setDetailModal: (props: DetailViewProps | null) => void;
  toggleDarkMode: () => void;
  appState: AppState;
  settings: AppSettings;
  navigate: {
    internal: (page: PageRoute | V2Routes, query?: any) => void;
    external: (url: string) => void;
  };
  notifySuccess: typeof notifySuccess;
  request: typeof window.fetch;
};

export const AppContext = React.createContext<AppContextType>(
  null as AppContextType
);

export interface AppProps {
  children?: any;
  renderFooter?: boolean;
  notifySuccess?: typeof notifySuccess;
}

export default function AppContextProvider({ ...props }: AppProps) {
  const history = useHistory();
  const [appState, setAppState] = React.useState({
    error: null,
    detailModal: null,
  });
  const [appSettings, setAppSettings] = React.useState<AppSettings>({
    renderFooter: props.renderFooter,
    theme:
      window.matchMedia("(prefers-color-scheme: dark)").matches ||
      localStorage.getItem("mode") === ThemeTypes.Dark
        ? ThemeTypes.Dark
        : ThemeTypes.Light,
  });

  const clearAsyncError = () => {
    setAppState({
      ...appState,
      error: null,
    });
  };

  React.useEffect(() => {
    // clear the error state on navigation
    clearAsyncError();
  }, [window.location]);

  const doAsyncError = (message: string, detail: string) => {
    console.error(message);
    setAppState({
      ...appState,
      error: { message, detail },
    });
  };

  const setDetailModal = (props: DetailViewProps | null) => {
    setAppState({ ...appState, detailModal: props });
  };

  const toggleDarkMode = () => {
    const newMode =
      appSettings.theme === ThemeTypes.Light
        ? ThemeTypes.Dark
        : ThemeTypes.Light;
    localStorage.setItem("mode", newMode);
    return setAppSettings({
      ...appSettings,
      theme: newMode,
    });
  };

  const value: AppContextType = {
    userConfigRepoName: "wego-github-jlw-config-repo",
    doAsyncError,
    clearAsyncError,
    setDetailModal,
    toggleDarkMode,
    appState,
    notifySuccess: props.notifySuccess || notifySuccess,
    settings: appSettings,
    navigate: {
      internal: (page: PageRoute, query?: any) => {
        const u = formatURL(page, query);

        history.push(u);
      },
      external: (url) => {
        if (process.env.NODE_ENV === "test") {
          return;
        }
        window.location.href = url;
      },
    },
    request: window.fetch,
  };

  return <AppContext.Provider {...props} value={value} />;
}
