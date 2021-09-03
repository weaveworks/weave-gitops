import _ from "lodash";
import { useEffect, useState } from "react";
import { useLocation } from "react-router-dom";

export const normalizePath = (pathname) => {
  return _.tail(pathname.split("/"));
};

export default function useNavigation(): {
  currentPage: string;
} {
  const location = useLocation();
  const [currentPage, setCurrentPage] = useState("");

  useEffect(() => {
    const [pageName] = normalizePath(location.pathname);
    setCurrentPage(pageName as string);
  }, [location]);

  return {
    currentPage,
  };
}
