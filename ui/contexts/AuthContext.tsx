import * as React from "react";
import LoadingPage from "../components/LoadingPage";
import { AuthSwitch } from "./AutoSwitch";

const USER_INFO = "oauth2/userinfo";

export type AuthContext = {
  signIn: (username?: string, password?: string) => void;
  userInfo: {
    email: string;
    groups: string[];
  };
};

export const Auth = React.createContext<AuthContext | null>(null);

export default function AuthContextProvider({ children }) {
  const [userInfo, setUserInfo] = React.useState<{
    email: string;
    groups: string[];
  } | null>(null);
  const [loading, setLoading] = React.useState<boolean>(false);

  const signIn = React.useCallback((username?: string, password?: string) => {
    const CURRENT_URL = window.origin;
    fetch(`/oauth2/sign_in?return_url=${encodeURIComponent(CURRENT_URL)}`, {
      method: "POST",
      body: JSON.stringify({ username, password }),
    })
      .then((res) => console.log(res))
      .catch((err) => console.log(err));
  }, []);

  const getUserInfo = React.useCallback(() => {
    setLoading(true);
    fetch(`/${USER_INFO}`)
      .then((response) => {
        return response.json();
      })
      .then((data) => setUserInfo({ email: data.email, groups: [] }))
      .catch((err) => {
        console.log(err);
        if (err.code === "401") {
          // user is not authenticated
        }
      })
      .finally(() => setLoading(false));
    // set state for user Info
    // if 401 => user not authenticated => leave null
  }, []);

  console.log(userInfo?.email);

  React.useEffect(() => {
    getUserInfo();
  }, [getUserInfo]);

  return (
    <Auth.Provider value={{ signIn, userInfo }}>
      {/* {loading ? <LoadingPage /> : null} */}
      {userInfo?.email && !loading ? children : <AuthSwitch />}
    </Auth.Provider>
  );
}
