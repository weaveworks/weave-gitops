import _ from "lodash";
import qs from "query-string";
import { useHistory, useLocation } from "react-router-dom";
import { PageRoute } from "../lib/types";
import { formatURL } from "../lib/utils";

export const normalizePath = (pathname) => {
  return _.tail(pathname.split("/"));
};

export default function useNavigation<T>(): {
  currentPage: string;
  query: T;
  navigate: (PageRoute, any?) => void;
} {
  const history = useHistory();
  const location = useLocation();
  const [pageName] = normalizePath(location.pathname);

  const navigate = (page: PageRoute, query: any) => {
    history.push(formatURL(page, query));
  };

  const query = qs.parse(location.search) as any;

  return {
    currentPage: pageName as string,
    query,
    navigate,
  };
}
