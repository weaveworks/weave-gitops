import * as React from "react";
import styled from "styled-components";
import { AppContext } from "../contexts/AppContext";
import { useDebounce, useRequestState } from "../hooks/common";
import {
  GitProvider,
  ParseRepoURLResponse,
} from "../lib/api/applications/applications.pb";
import Button from "./Button";
import Flex from "./Flex";
import Input, { InputProps } from "./Input";

type Props = InputProps & {
  onAuthClick: (provider: GitProvider) => void;
};

function RepoInputWithAuth(props: Props) {
  const { applicationsClient } = React.useContext(AppContext);
  const [res, , err, req] = useRequestState<ParseRepoURLResponse>();

  const debouncedURL = useDebounce<string>(props.value as string, 500);

  React.useEffect(() => {
    if (!debouncedURL) {
      return;
    }
    req(applicationsClient.ParseRepoURL({ url: debouncedURL }));
  }, [debouncedURL]);

  return (
    <Flex className={props.className} align>
      <Input
        {...props}
        error={props.value && !!err?.message}
        helperText={!props.value || !err ? props.helperText : err?.message}
      />

      <Button
        id="auth-button"
        variant="contained"
        className={res?.provider}
        disabled={!res?.provider}
        onClick={() => props.onAuthClick(res.provider)}
      >
        {res?.provider
          ? `Authenticate with ${res.provider}`
          : "Authenticate with your Git Provider"}
      </Button>
    </Flex>
  );
}

export default styled(RepoInputWithAuth).attrs({
  className: RepoInputWithAuth.name,
})`
  #auth-button {
    margin-left: 8px;
  }

  .GitHub {
    background-color: black;
    color: white;
  }

  .GitLab {
    background-color: #fc6d26;
    color: white;
  }
`;
