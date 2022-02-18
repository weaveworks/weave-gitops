import * as React from "react";
import LoadingPage from "../components/LoadingPage";
import { AuthSwitch } from "./AutoSwitch";
import { useHistory } from "react-router-dom";

const USER_INFO = "/oauth2/userinfo";
const SIGN_IN = "/oauth2/sign_in";

const Loader: React.FC<{ loading?: boolean }> = ({
  children,
  loading = true,
  ...props
}) => {
  console.log("loader");
  return <>{loading ? <LoadingPage /> : children}</>;
};

export type AuthContext = {
  signIn: (data: any) => void;
  userInfo: {
    email: string;
    groups: string[];
  };
};

export const Auth = React.createContext<AuthContext | null>(null);

export default function AuthContextProvider({ children }) {
  const [userInfo, setUserInfo] = React.useState<
    | {
        email: string;
        groups: string[];
      }
    | undefined
  >(undefined);
  const [loading, setLoading] = React.useState<boolean>(false);
  const [authenticated, setAuthenticated] = React.useState<boolean>();
  const history = useHistory();
  const {
    location: { pathname },
  } = history;

  const signIn = React.useCallback((data) => {
    setLoading(true);
    fetch(SIGN_IN, {
      method: "POST",
      body: JSON.stringify(data),
    })
      .then((response) => {
        if (response.status === 200) {
          setAuthenticated(true);
        }
      })
      .catch((err) => console.log(err))
      .finally(() => setLoading(false));
  }, []);

  const getUserInfo = React.useCallback(() => {
    setLoading(true);
    fetch(USER_INFO)
      .then((response) => {
        return response.json();
      })
      .then((data) => {
        setUserInfo({ email: data.email, groups: [] });

        console.log(pathname);
        if (pathname === "/sign_in") {
          history.push("/");
        }
      })
      .catch((err) => setUserInfo(undefined))
      .finally(() => setLoading(false));
  }, []);

  React.useEffect(() => {
    getUserInfo();
    return history.listen(getUserInfo);
  }, [getUserInfo, history]);

  console.log("email", userInfo?.email);
  console.log("path", pathname);
  console.log("authenticated", authenticated);
  console.log("history", history);

  return (
    <Auth.Provider
      value={{
        signIn,
        userInfo,
      }}
    >
      <Loader loading={loading}>
        {userInfo?.email !== undefined ? children : <AuthSwitch />}
      </Loader>
    </Auth.Provider>
  );
}
