import * as React from "react";
import { Redirect, Route, RouteProps } from "react-router-dom";
import useAuth from "../hooks/auth";
import useNavigation from "../hooks/navigation";
import { PageRoute } from "../lib/types";
import LoadingPage from "./LoadingPage";

export default function AuthenticatedRoute({
  component: Component,
  ...rest
}: RouteProps) {
  const { navigate, currentPage } = useNavigation();
  const { user, loading } = useAuth();

  React.useEffect(() => {
    if (!loading && !user) {
      navigate(PageRoute.Auth, { next: currentPage });
    }
  }, []);

  if (loading) {
    return <LoadingPage />;
  }

  return (
    <Route
      {...rest}
      component={(p) =>
        user ? <Component {...p} /> : <Redirect to={PageRoute.Auth} />
      }
    />
  );
}
