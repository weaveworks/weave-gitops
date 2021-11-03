// eslint-disable-next-line
import { ButtonProps } from "@material-ui/core";
import * as React from "react";
import styled from "styled-components";
import { AppContext } from "../contexts/AppContext";
import useNavigation from "../hooks/navigation";
import { CallbackSessionState } from "../lib/storage";
import { PageRoute } from "../lib/types";
import { gitlabOAuthRedirectURI } from "../lib/utils";
import Button from "./Button";

type Props = { callbackState?: CallbackSessionState } & ButtonProps;

function GitlabAuthButton({ callbackState, ...props }: Props) {
  const { currentPage } = useNavigation();
  const { applicationsClient, storeCallbackState, navigate } =
    React.useContext(AppContext);

  const handleClick = () => {
    applicationsClient
      .GetGitlabAuthURL({
        redirectUri: gitlabOAuthRedirectURI(),
      })
      .then((res) => {
        if (!callbackState) {
          callbackState = { page: `/${currentPage}` as PageRoute, state: null };
        }
        storeCallbackState(callbackState);
        navigate(res.url);
      });
  };

  return (
    <Button {...props} variant="contained" onClick={handleClick}>
      Authenticate with GitLab
    </Button>
  );
}

export default styled(GitlabAuthButton).attrs({
  className: GitlabAuthButton.name,
})`
  &.MuiButton-contained {
    background-color: #fc6d26;
    color: white;
  }
`;
