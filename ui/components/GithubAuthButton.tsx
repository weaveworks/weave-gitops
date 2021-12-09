// eslint-disable-next-line
import { ButtonProps } from "@material-ui/core";
import * as React from "react";
import styled from "styled-components";
import { GithubAuthContext } from "../contexts/GithubAuthContext";
import { GitProvider } from "../lib/api/applications/applications.pb";
import Button from "./Button";

type Props = ButtonProps;

function GithubAuthButton(props: Props) {
  return <Button {...props}>Authenticate with {GitProvider.GitHub}</Button>;
}

export default styled(GithubAuthButton).attrs({
  className: GithubAuthButton.name,
})`
  &.MuiButton-contained {
    background-color: black;
    color: white;
  }
`;

export function GlobalGithubAuthButton({ onSuccess }) {
  const {
    dialogState: { success },
    setDialogState,
  } = React.useContext(GithubAuthContext);

  React.useEffect(() => {
    if (success && onSuccess) {
      onSuccess();
    }
  }, [success]);

  return <GithubAuthButton onClick={() => setDialogState(true, "")} />;
}
