import * as React from "react";
import { useNavigate } from "react-router-dom";
import { DetailViewProps } from "../components/DetailModal";
import { formatURL } from "../lib/nav";
import { PageRoute, V2Routes } from "../lib/types";
import { notifySuccess } from "../lib/utils";

type AppState = {
  error: null | { fatal: boolean; message: string; detail?: string };
  detailModal: DetailViewProps;
};

type AppSettings = {
  renderFooter: boolean;
};

export type AppContextType = {
  userConfigRepoName: string;
  doAsyncError: (message: string, detail: string) => void;
  clearAsyncError: () => void;
  setDetailModal: (props: DetailViewProps | null) => void;
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
  const navigate = useNavigate();
  const [appState, setAppState] = React.useState({
    error: null,
    detailModal: null,
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

  const value: AppContextType = {
    userConfigRepoName: "wego-github-jlw-config-repo",
    doAsyncError,
    clearAsyncError,
    setDetailModal,
    appState,
    notifySuccess: props.notifySuccess || notifySuccess,
    settings: {
      renderFooter: props.renderFooter,
    },
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
    request: window.fetch,
  };

  return <AppContext.Provider {...props} value={value} />;
}
