import * as React from "react";
import { useHistory } from "react-router-dom";
import Layout from "../components/Layout";
import LoadingPage from "../components/LoadingPage";
import { AuthSwitch } from "./AutoSwitch";

const USER_INFO = "/oauth2/userinfo";
const SIGN_IN = "/oauth2/sign_in";

const Loader: React.FC<{ loading?: boolean }> = ({ children, loading }) => {
  const history = useHistory();
  const {
    location: { pathname },
  } = history;
  return (
    <>
      {loading && pathname !== "/sign_in" ? (
        <Layout>
          <LoadingPage />
        </Layout>
      ) : (
        children
      )}
    </>
  );
};

export type AuthContext = {
  signIn: (data: any) => void;
  userInfo: {
    email: string;
    groups: string[];
  };
  error: { status: number; statusText: string };
  loading: boolean;
};

export const Auth = React.createContext<AuthContext | null>(null);

export default function AuthContextProvider({ children }) {
  const [userInfo, setUserInfo] = React.useState<{
    email: string;
    groups: string[];
  }>(null);
  const [loading, setLoading] = React.useState<boolean>(true);
  const [error, setError] = React.useState(null);
  const history = useHistory();

  const signIn = React.useCallback((data) => {
    setLoading(true);
    fetch(SIGN_IN, {
      method: "POST",
      body: JSON.stringify(data),
    })
      .then((response) => {
        if (response.status !== 200) {
          setError(response);
          return;
        }
        getUserInfo().then(() => history.push("/"));
      })
      .finally(() => setLoading(false));
  }, []);

  const getUserInfo = React.useCallback(() => {
    setLoading(true);
    return fetch(USER_INFO)
      .then((response) => {
        if (response.status === 400 || response.status === 401) {
          setUserInfo(null);
          return;
        }
        return response.json();
      })
      .then((data) => setUserInfo({ email: data?.email, groups: [] }))
      .catch((err) => console.log(err))
      .finally(() => setLoading(false));
  }, []);

  React.useEffect(() => {
    getUserInfo();
    return history.listen(getUserInfo);
  }, [getUserInfo, history]);

  return (
    <Auth.Provider
      value={{
        signIn,
        userInfo,
        error,
        loading,
      }}
    >
      <Loader loading={loading}>
        {userInfo?.email ? children : <AuthSwitch />}
      </Loader>
    </Auth.Provider>
  );
}
