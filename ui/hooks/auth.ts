import { useContext, useState } from "react";
import { AppContext } from "../contexts/AppContext";
import { OauthProvider } from "../lib/api/applications/applications.pb";
import { storeToken } from "../lib/storage";

export default function useAuth() {
  const { applicationsClient, doAsyncError, user, loading } =
    useContext(AppContext);
  const [providers, setProviders] = useState<OauthProvider[]>([]);

  const authenticate = (code: string, providerName: string) =>
    applicationsClient
      .Authenticate({ code, providerName })
      .then((res) => {
        storeToken(res.token);
      })
      .catch((err) => doAsyncError("Could not authenticate", err.message));

  const getProviders = () =>
    applicationsClient
      .GetAuthenticationProviders({})
      .then((res) => setProviders(res.providers))
      .catch((err) =>
        doAsyncError("Could not get auth providers", err.message)
      );

  return {
    user,
    loading,
    providers,
    authenticate,
    getProviders,
  };
}
