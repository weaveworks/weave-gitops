import * as React from "react";
import { RequestError } from "../lib/types";
import Alert from "./Alert";
import LoadingPage from "./LoadingPage";
import Spacer from "./Spacer";

type Props = {
  className?: string;
  loading?: boolean;
  error: RequestError;
  children: any;
};

function RequestStateHandler({ loading, error, children }: Props) {
  if (loading) {
    return <LoadingPage />;
  }

  if (error) {
    return (
      <Spacer padding="small">
        <Alert severity="error" message={error.message} />
      </Spacer>
    );
  }

  return children;
}

export default RequestStateHandler;
