import * as React from "react";
import LoadingPage from "../components/LoadingPage";
import { AuthSwitch } from "./AutoSwitch";
import { useHistory } from "react-router-dom";
import SignIn from "../pages/SignIn";

const USER_INFO = "/oauth2/userinfo";
const SIGN_IN = "/oauth2/sign_in";

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

  const signIn = React.useCallback((data) => {
    fetch(SIGN_IN, {
      method: "POST",
      body: JSON.stringify(data),
    })
      .then((response) => {
        if (response.status === 200) {
          setAuthenticated(true);
        }
      })
      .catch((err) => console.log(err));
  }, []);

  const getUserInfo = React.useCallback(() => {
    setLoading(true);
    fetch(USER_INFO)
      .then((response) => {
        return response.json();
      })
      .then((data) => {
        setUserInfo({ email: data.email, groups: [] });
        // this should only happen at sign in
        history.push("/");
      })
      .catch((err) => setUserInfo(undefined))
      .finally(() => setLoading(false));
  }, []);

  React.useEffect(() => {
    getUserInfo();
  }, [authenticated, getUserInfo, window.location]);

  console.log(userInfo?.email);
  console.log(window.location.pathname);
  console.log(authenticated);

  return (
    <Auth.Provider value={{ signIn, userInfo }}>
      {/* {loading ? <LoadingPage /> : null} */}
      {userInfo?.email !== undefined ? children : <AuthSwitch />}
    </Auth.Provider>
  );
}
