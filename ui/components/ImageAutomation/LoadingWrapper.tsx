import React from "react";
import LoadingPage from "../LoadingPage";
import { Errors, PageProps } from "../Page";

const LoadingWrapper = ({ loading, error, children }: PageProps) => {
  return (
    <>
      {loading && <LoadingPage />}
      {error && <Errors error={error} />}
      {!loading && !error && children}
    </>
  );
};

export default LoadingWrapper;
