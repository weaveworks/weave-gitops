import _ from "lodash";
import { useContext } from "react";
import { useLocation } from "react-router-dom";
import { AppContext, AppContextType } from "../contexts/AppContext";

export const normalizePath = (pathname) => {
  return _.tail(pathname.split("/"));
};

export default function useNavigation(): {
  currentPage: string;
  navigate: AppContextType["navigate"];
} {
  const { navigate } = useContext(AppContext);
  const location = useLocation();

  return {
    currentPage: location.pathname,
    navigate,
  };
}
