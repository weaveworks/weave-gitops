import { Button } from "@material-ui/core";
import * as React from "react";
import { DefaultGitOps } from "../lib/rpc/gitops";
import { wrappedFetch } from "../lib/util";

const gitopsClient = new DefaultGitOps("/api/gitops", wrappedFetch);

function Login() {
  const handleLogin = () => {
    gitopsClient.login({ state: "" }).then((res) => {
      window.location.href = res.redirectUrl;
    });

    gitopsClient
      .addApplication({
        name: "my-app",
        deploymentType: "kustomize",
      })
      .then((res) => console.log(res.application.name));
  };
  return (
    <Button onClick={handleLogin} color="primary">
      Login
    </Button>
  );
}

export default Login;
