import * as React from "react";
import { Applications } from "../lib/api/applications/applications.pb";

export type AppContextType = {
  applicationsClient: typeof Applications;
};

export const AppContext = React.createContext<AppContextType>(null as any);

export default function AppContextProvider({ applicationsClient, ...props }) {
  const value: AppContextType = {
    applicationsClient,
  };

  return <AppContext.Provider {...props} value={value} />;
}
