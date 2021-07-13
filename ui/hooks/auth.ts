import { useContext, useState } from "react";
import { AppContext } from "../contexts/AppContext";
import { OauthProvider } from "../lib/api/applications/applications.pb";
import { removeToken, storeToken } from "../lib/storage";
import { AsyncError, PageRoute } from "../lib/types";
import useNavigation from "./navigation";

export default function useAuth() {
  const {
    applicationsClient,
    user,
    loading: appLoading,
    setUser,
  } = useContext(AppContext);
  const { navigate, query } = useNavigation();

  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<AsyncError>(null);
  const [providers, setProviders] = useState<OauthProvider[]>([]);

  const authenticate = (code: string, providerName: string) => {
    setLoading(true);

    return applicationsClient
      .Authenticate({ code, providerName })
      .then((res) => {
        storeToken(res.token);
        return res.token;
      })
      .then((token) => {
        applicationsClient.GetUser({ token }).then(({ user }) => {
          setUser(user);
        });
      })
      .catch((err) =>
        setError({
          message: "Could not authenticate",
          detail: err.message,
        })
      )
      .finally(() => setLoading(false));
  };

  const getProviders = () => {
    setLoading(true);

    return applicationsClient
      .GetAuthenticationProviders({})
      .then((res) => setProviders(res.providers))
      .catch((err) =>
        setError({
          message: "Could not get auth providers",
          detail: err.message,
        })
      )
      .finally(() => setLoading(false));
  };

  const logout = () => {
    removeToken();
    setUser(null);
    navigate(PageRoute.Auth);
  };

  return {
    user,
    error,
    loading: appLoading || loading,
    providers,
    authenticate,
    getProviders,
    logout,
  };
}
