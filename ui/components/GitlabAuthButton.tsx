// eslint-disable-next-line
import { ButtonProps } from "@material-ui/core";
import * as React from "react";
import styled from "styled-components";
import { AppContext } from "../contexts/AppContext";
import { gitlabOAuthRedirectURI } from "../lib/utils";
import Button from "./Button";

type Props = ButtonProps;

function GitlabAuthButton({ ...props }: Props) {
  const { applicationsClient, navigate } = React.useContext(AppContext);

  const handleClick = (e) => {
    if (props.onClick) {
      props.onClick(e);
    }
    applicationsClient
      .GetGitlabAuthURL({
        redirectUri: gitlabOAuthRedirectURI(),
      })
      .then((res) => {
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
