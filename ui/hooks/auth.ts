import { useContext, useEffect, useState } from "react";
import { AppContext } from "../contexts/AppContext";
import {
  GetGithubAuthStatusResponse,
  GetGithubDeviceCodeResponse,
  GitProvider,
} from "../lib/api/applications/applications.pb";
import { GrpcErrorCodes } from "../lib/types";
import { poller } from "../lib/utils";

export function useIsAuthenticated(provider: GitProvider): boolean {
  const [isAuthenticated, setIsAuthenticated] = useState(false);
  const { getProviderToken } = useContext(AppContext);

  useEffect(() => {
    const token = getProviderToken(provider);

    if (token) {
      setIsAuthenticated(true);
    }
  }, [provider]);

  return isAuthenticated;
}

export default function useAuth() {
  const [loading, setLoading] = useState(true);
  const { applicationsClient, getProviderToken, storeProviderToken } =
    useContext(AppContext);

  const getGithubDeviceCode = () => {
    setLoading(true);
    return applicationsClient
      .GetGithubDeviceCode({})
      .finally(() => setLoading(false));
  };

  const getGithubAuthStatus = (codeRes: GetGithubDeviceCodeResponse) => {
    let poll;
    return {
      cancel: () => clearInterval(poll),
      promise: new Promise<GetGithubAuthStatusResponse>((accept, reject) => {
        poll = poller(() => {
          applicationsClient
            .GetGithubAuthStatus(codeRes)
            .then((res) => {
              clearInterval(poll);
              accept(res);
            })
            .catch(({ code, message }) => {
              // Unauthenticated means we can keep polling.
              //  On anything else, stop polling and report.
              if (code !== GrpcErrorCodes.Unauthenticated) {
                clearInterval(poll);
                reject({ message });
              }
            });
        }, (codeRes.interval + 1) * 1000);
      }),
    };
  };

  return {
    loading,
    getGithubDeviceCode,
    getGithubAuthStatus,
    getProviderToken,
    storeProviderToken,
  };
}
