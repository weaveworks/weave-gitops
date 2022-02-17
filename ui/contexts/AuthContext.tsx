import * as React from "react";
import LoadingPage from "../components/LoadingPage";
import { AuthSwitch } from "./AutoSwitch";
import { useHistory } from "react-router-dom";

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
  const [userInfo, setUserInfo] = React.useState<{
    email: string;
    groups: string[];
  } | null>(null);
  const [loading, setLoading] = React.useState<boolean>(false);
  const [auth, setAuth] = React.useState<boolean>(false);
  const history = useHistory();

  const signIn = React.useCallback((data) => {
    fetch(SIGN_IN, {
      method: "POST",
      body: JSON.stringify(data),
    })
      .then((response) => {
        if (response.status === 200) {
          setAuth(true);
        }
      })
      .catch((err) => console.log(err));
  }, []);

  const getUserInfo = React.useCallback(() => {
    // setLoading(true);
    fetch(USER_INFO)
      .then((response) => {
        return response.json();
      })
      .then((data) => setUserInfo({ email: data.email, groups: [] }))
      .catch((err) => {
        console.log(err);
        if (err.code === "401") {
          setUserInfo(null);
        }
      });
    // .finally(() => setLoading(false));
  }, []);

  React.useEffect(() => {
    getUserInfo();
  }, [auth, getUserInfo]);

  return (
    <Auth.Provider value={{ signIn, userInfo }}>
      {/* {loading ? <LoadingPage /> : null} */}
      {userInfo?.email && !loading ? children : <AuthSwitch />}
    </Auth.Provider>
  );
}
