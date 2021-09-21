import qs from "query-string";
import { toast } from "react-toastify";
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

export function notifySuccess(message: string) {
  toast["success"](message);
}

export function notifyError(message: string) {
  toast["error"](`Error: ${message}`);
}

const tokenKey = (providerName: string) => `gitProviderToken_${providerName}`;

export function storeProviderToken(providerName: string, token: string) {
  if (!window.localStorage) {
    console.warn("no local storage found");
    return;
  }

  localStorage.setItem(tokenKey(providerName), token);
}

export function getProviderToken(providerName: string): string {
  if (!window.localStorage) {
    console.warn("no local storage found");
    return;
  }

  return localStorage.getItem(tokenKey(providerName));
}
