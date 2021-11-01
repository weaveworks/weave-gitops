import { CircularProgress } from "@material-ui/core";
import * as React from "react";
import styled from "styled-components";
import { AppContext } from "../contexts/AppContext";
import { useDebounce, useRequestState } from "../hooks/common";
import { ParseRepoURLResponse } from "../lib/api/applications/applications.pb";
import Flex from "./Flex";
import Input, { InputProps } from "./Input";

function RepoInputWithAuth(props: InputProps) {
  const { applicationsClient } = React.useContext(AppContext);
  const [res, loading, err, req] = useRequestState<ParseRepoURLResponse>();

  const debouncedURL = useDebounce<string>(props.value as string, 500);

  React.useEffect(() => {
    if (!debouncedURL) {
      return;
    }
    req(applicationsClient.ParseRepoURL({ url: debouncedURL }));
  }, [debouncedURL]);

  return (
    <Flex>
      <Input {...props} />
      <div>{loading && <CircularProgress />}</div>
      <div>{!loading && res && `Provider: ${res.provider}`}</div>
      <div>{err && `Error: ${err.message}`}</div>
    </Flex>
  );
}

export default styled(RepoInputWithAuth).attrs({
  className: RepoInputWithAuth.name,
})``;
