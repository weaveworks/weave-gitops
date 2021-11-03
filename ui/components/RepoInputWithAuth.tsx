// eslint-disable-next-line
import * as React from "react";
import styled from "styled-components";
import { AppContext } from "../contexts/AppContext";
import { useDebounce, useRequestState } from "../hooks/common";
import {
  GitProvider,
  ParseRepoURLResponse,
} from "../lib/api/applications/applications.pb";
import { CallbackSessionState } from "../lib/storage";
import Button from "./Button";
import Flex from "./Flex";
import GithubAuthButton from "./GithubAuthButton";
import GitlabAuthButton from "./GitlabAuthButton";
import Icon, { IconType } from "./Icon";
import Input, { InputProps } from "./Input";

type Props = InputProps & {
  onAuthClick?: (provider: GitProvider) => void;
  callbackState: CallbackSessionState;
  onProviderChange?: (provider: GitProvider) => void;
  isAuthenticated?: boolean;
};

function RepoInputWithAuth({
  onAuthClick,
  onProviderChange,
  callbackState,
  isAuthenticated,
  ...props
}: Props) {
  const { applicationsClient } = React.useContext(AppContext);
  const [res, , err, req] = useRequestState<ParseRepoURLResponse>();
  const debouncedURL = useDebounce<string>(props.value as string, 500);

  React.useEffect(() => {
    if (!debouncedURL) {
      return;
    }
    req(applicationsClient.ParseRepoURL({ url: debouncedURL }));
  }, [debouncedURL]);

  React.useEffect(() => {
    if (!res) {
      return;
    }

    if (res.provider && onProviderChange) {
      onProviderChange(res.provider);
    }
  }, [res]);

  const AuthButton =
    res?.provider === GitProvider.GitHub ? (
      <GithubAuthButton
        onClick={() => {
          onAuthClick(GitProvider.GitHub);
        }}
      />
    ) : (
      <GitlabAuthButton callbackState={callbackState} />
    );

  const renderProviderAuthButton =
    props.value && !!res?.provider && !isAuthenticated;

  return (
    <Flex className={props.className} align>
      <Input
        {...props}
        error={props.value && !!err?.message ? true : false}
        helperText={!props.value || !err ? props.helperText : err?.message}
      />
      <div className="auth-message">
        {isAuthenticated && (
          <Flex align>
            <Icon size="medium" color="success" type={IconType.CheckMark} />{" "}
            {res?.provider} credentials detected
          </Flex>
        )}
        {!isAuthenticated && !res && (
          <Button disabled variant="contained">
            Authenticate with your Git Provider
          </Button>
        )}

        {renderProviderAuthButton ? AuthButton : null}
      </div>
    </Flex>
  );
}

export default styled(RepoInputWithAuth).attrs({
  className: RepoInputWithAuth.name,
})`
  .auth-message {
    margin-left: 8px;

    ${Icon} {
      margin-right: 4px;
    }
  }
`;
