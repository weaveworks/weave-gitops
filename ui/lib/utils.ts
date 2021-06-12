import qs from "query-string";
import { PageRoute } from "./types";

export const formatURL = (page: string, query: any = {}) => {
  return `${page}?${qs.stringify(query)}`;
};

export const getNavValue = (currentPage: any): PageRoute | boolean => {
  switch (currentPage) {
    case "applications":
    case "application_detail":
      return PageRoute.Applications;

    default:
      // The "Tabs" component of material-ui wants a bool
      return false;
  }
};
