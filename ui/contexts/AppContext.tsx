import * as React from "react";
import type { JSX } from "react";
import { useNavigate } from "react-router";
import { DetailViewProps } from "../components/DetailModal";
import { formatURL } from "../lib/nav";
import { PageRoute, V2Routes } from "../lib/types";
import { notifySuccess, withBasePath } from "../lib/utils";

type AppState = {
  error: null | { fatal: boolean; message: string; detail?: string };
  detailModal: DetailViewProps;
};

export enum ThemeTypes {
  Light = "light",
  Dark = "dark",
}

type AppSettings = {
  footer: JSX.Element;
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
  null as AppContextType,
);

export interface AppProps {
  children?: any;
  footer?: JSX.Element;
  notifySuccess?: typeof notifySuccess;
}

export default function AppContextProvider({ ...props }: AppProps) {
  const navigate = useNavigate();
  const [appState, setAppState] = React.useState({
    error: null,
    detailModal: null,
  });
  const [appSettings, setAppSettings] = React.useState<AppSettings>({
    footer: props.footer,
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
        navigate(u);
      },
      external: (url) => {
        if (process.env.NODE_ENV === "test") {
          return;
        }
        window.location.href = url;
      },
    },
    request: (
      input: RequestInfo | URL,
      init?: RequestInit,
    ): Promise<Response> => {
      if (typeof input === "string") {
        input = withBasePath(input);
      }
      return window.fetch(input, init);
    },
  };

  return <AppContext.Provider {...props} value={value} />;
}
