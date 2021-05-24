import * as React from "react";

export default function App() {
  const [ok, setOk] = React.useState(false);
  const [error, setError] = React.useState(null);

  React.useEffect(() => {
    fetch("/api")
      .then((res) => res.json())
      .then((res) => {
        setOk(res);
      })
      .catch((err) => {
        setError(err);
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
