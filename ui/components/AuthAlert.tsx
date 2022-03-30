import * as React from "react";
import styled from "styled-components";
import { GitProvider } from "../lib/api/gitauth/gitauth.pb";
import { RequestError } from "../lib/types";
import Alert from "./Alert";
import Button from "./Button";
import Flex from "./Flex";
import GithubAuthButton from "./GithubAuthButton";
import GitlabAuthButton from "./GitlabAuthButton";

type Props = {
  className?: string;
  error?: RequestError;
  onClick: () => void;
  title: string;
  provider: GitProvider;
};

const Message = ({ onClick, provider }) => (
  <Flex align wide between>
    Could not authenticate with your Git Provider{" "}
    {provider === GitProvider.GitHub ? (
      <GithubAuthButton onClick={onClick} />
    ) : (
      <GitlabAuthButton onClick={onClick} />
    )}
  </Flex>
);

function AuthAlert({ className, onClick, title, provider }: Props) {
  return (
    <Alert
      className={className}
      severity="error"
      title={title}
      message={<Message provider={provider} onClick={onClick} />}
    />
  );
}

export default styled(AuthAlert).attrs({ className: AuthAlert.name })`
  ${Button} {
    margin-left: 16px;
  }
`;
