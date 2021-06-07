import * as React from "react";

const AuthContext = React.createContext(null);

export function useAuth() {
  return React.useContext(AuthContext);
}

export function AuthProvider({ children }: any) {
  const [currentUser, setCurrentUser] = React.useState<any>();
  const [loading, setLoading] = React.useState(true);

  React.useEffect(() => {
    fetch("/api/authorize")
      .then((res) => res.json())
      .then((res) => {
        setCurrentUser(res);
      })
      .finally(() => {
        setLoading(false);
      });
  }, []);

  const value = {
    currentUser,
  };

  if (loading) {
    return <div>Loading...</div>;
  }

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}
