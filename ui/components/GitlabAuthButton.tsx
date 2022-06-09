// eslint-disable-next-line
import { ButtonProps } from "@material-ui/core";
import * as React from "react";
import styled from "styled-components";
import { AppContext } from "../contexts/AppContext";
import { CallbackStateContext } from "../contexts/CallbackStateContext";
import { gitlabOAuthRedirectURI } from "../lib/utils";
import Button from "./Button";

type Props = ButtonProps;

function GitlabAuthButton({ ...props }: Props) {
  const { callbackState } = React.useContext(CallbackStateContext);
  const { applicationsClient, navigate, storeCallbackState } =
    React.useContext(AppContext);

  const handleClick = (e) => {
    e.preventDefault();

    storeCallbackState(callbackState);

    applicationsClient
      .GetGitlabAuthURL({
        redirectUri: gitlabOAuthRedirectURI(),
      })
      .then((res) => {
        navigate.external(res.url);
      });
  };

  return (
    <Button {...props} onClick={handleClick}>
      Authenticate with GitLab
    </Button>
  );
}

export default styled(GitlabAuthButton).attrs({
  className: GitlabAuthButton.name,
})``;
