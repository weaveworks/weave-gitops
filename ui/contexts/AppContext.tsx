import * as React from "react";
import { useHistory } from "react-router-dom";
import { formatURL } from "../lib/nav";
import { FluxObject, FluxObjectNode } from "../lib/objects";
import { PageRoute, V2Routes } from "../lib/types";
import { notifySuccess } from "../lib/utils";

type AppState = {
  error: null | { fatal: boolean; message: string; detail?: string };
  nodeYaml: FluxObjectNode | null;
};

type AppSettings = {
  renderFooter: boolean;
  theme: "light" | "dark";
};

export type AppContextType = {
  userConfigRepoName: string;
  doAsyncError: (message: string, detail: string) => void;
  clearAsyncError: () => void;
  setNodeYaml: (obj: FluxObject | FluxObjectNode) => void;
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
    nodeYaml: null,
  });
  const [appSettings, setAppSettings] = React.useState<AppSettings>({
    renderFooter: props.renderFooter,
    theme:
      window.matchMedia("prefers-color-scheme: dark").matches ||
      localStorage.getItem("mode") === "dark"
        ? "dark"
        : "light",
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

  const setNodeYaml = (obj: FluxObject | FluxObjectNode) => {
    if (obj) setAppState({ ...appState, nodeYaml: obj });
    else setAppState({ ...appState, nodeYaml: null });
  };

  const toggleDarkMode = () => {
    const newMode = appSettings.theme === "light" ? "dark" : "light";
    localStorage.setItem("mode", newMode);
    setAppSettings({
      ...appSettings,
      theme: newMode,
    });
  };

  const value: AppContextType = {
    userConfigRepoName: "wego-github-jlw-config-repo",
    doAsyncError,
    clearAsyncError,
    setNodeYaml,
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
