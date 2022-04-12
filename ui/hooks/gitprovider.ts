import _ from "lodash";
import { useCallback, useContext, useState } from "react";
import { AppContext } from "../contexts/AppContext";
import {
  GetGithubAuthStatusResponse,
  GetGithubDeviceCodeResponse,
  GitProvider,
  ValidateProviderTokenResponse,
} from "../lib/api/applications/applications.pb";
import { GrpcErrorCodes } from "../lib/types";
import { poller } from "../lib/utils";
import { makeHeaders, useRequestState } from "./common";

export function useIsAuthenticated() {
  const [res, loading, error, req] =
    useRequestState<ValidateProviderTokenResponse>();

  const { getProviderToken, applicationsClient } = useContext(AppContext);

  return {
    isAuthenticated: error ? false : res?.valid,
    loading,
    error,
    req: useCallback(
      (provider: GitProvider) => {
        const headers = makeHeaders(_.bind(getProviderToken, this, provider));
        req(
          applicationsClient.ValidateProviderToken({ provider }, { headers })
        );
      },
      [makeHeaders, getProviderToken, applicationsClient.ValidateProviderToken]
    ),
  };
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
