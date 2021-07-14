import { useContext, useEffect, useState } from "react";
import { AppContext } from "../contexts/AppContext";
import { OauthProvider, User } from "../lib/api/applications/applications.pb";
import { storeToken } from "../lib/storage";

export default function useAuth() {
  const { applicationsClient, doAsyncError } = useContext(AppContext);
  const [loading, setLoading] = useState(true);
  const [providers, setProviders] = useState<OauthProvider[]>([]);
  const [user, setUser] = useState<User>();

  useEffect(() => {
    applicationsClient
      .GetUser({})
      .then((res) => {
        setUser(res.user);
      })
      .catch((err) => doAsyncError("Error getting user", err.message))
      .finally(() => setLoading(false));
  }, []);

  const authenticate = (code: string, providerName: string) =>
    applicationsClient
      .Authenticate({ code, providerName })
      .then((res) => {
        storeToken(res.token);
        setUser(res.user);
      })
      .catch((err) => doAsyncError("Could not authenticate", err.message));

  const getProviders = () =>
    applicationsClient
      .GetAuthenticationProviders({})
      .then((res) => setProviders(res.providers))
      .catch((err) =>
        doAsyncError("Could not get auth providers", err.message)
      );

  const getUser = () =>
    applicationsClient.GetUser({}).then((res) => setUser(res.user));

  return {
    user,
    loading,
    providers,
    authenticate,
    getUser,
    getProviders,
  };
}
