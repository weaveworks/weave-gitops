import * as React from "react";
import { DefaultGitOps } from "./lib/rpc/gitops";
import { wrappedFetch } from "./lib/util";

const gitopsClient = new DefaultGitOps("/api/gitops", wrappedFetch);

export default function App() {
  const [ok, setOk] = React.useState(false);
  const [error, setError] = React.useState(null);

  // React.useEffect(() => {
  //   fetch("/api")
  //     .then((res) => res.json())
  //     .then((res) => {
  //       setOk(res);
  //     })
  //     .catch((err) => {
  //       setError(err);
  //     });
  // }, []);

  React.useEffect(() => {
    gitopsClient.listApplications({}).then((res) => {
      console.log(res);
    });
  }, []);

  return (
    <div>
      Weave GitOps UI
      <p>API Server Reponse:</p>
      <pre>{JSON.stringify(ok)}</pre>
      {error && <p>{error.msg}</p>}
    </div>
  );
}
