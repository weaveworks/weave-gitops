import * as React from "react";

type Props = {
  children?: any;
};

export interface GithubAuthContextType {
  dialogState: { open: boolean; repoName: string; success: boolean };
  setDialogState: (open: boolean, repoName: string) => void;
  setSuccess: () => void;
}

export const GithubAuthContext =
  React.createContext<GithubAuthContextType>(null);

export default function GithubAuthContextProvider({ children }: Props) {
  const [dialogState, setDialogState] = React.useState({
    open: false,
    repoName: null,
    success: false,
  });

  const value: GithubAuthContextType = {
    dialogState,
    setDialogState: (open: boolean, repoName: string) =>
      setDialogState({ ...dialogState, open, repoName }),
    setSuccess: () => setDialogState({ ...dialogState, success: true }),
  };
  return <GithubAuthContext.Provider value={value} children={children} />;
}
