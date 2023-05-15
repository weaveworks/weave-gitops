import React from "react";
import { AppContext, ThemeTypes } from "../contexts/AppContext";

export const useInDarkMode = (): boolean => {
  const { settings } = React.useContext(AppContext);
  return settings.theme === ThemeTypes.Dark;
};
