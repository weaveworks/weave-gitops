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

// Must be one of the valid URLs that we have already
// configured on the Gitlab backend for our Oauth app.
export function gitlabOAuthRedirectURI() {
  return `${window.location.origin}${PageRoute.GitlabOAuthCallback}`;
}

export function poller(cb, interval) {
  if (process.env.NODE_ENV === "test") {
    // Stay synchronous in tests
    return cb();
  }

  return setInterval(cb, interval);
}
