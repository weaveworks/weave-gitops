import { CircularProgress } from "@material-ui/core";
import * as React from "react";
import { useHistory } from "react-router-dom";
import styled from "styled-components";
import Alert from "../components/Alert";
import Flex from "../components/Flex";
import Page from "../components/Page";
import { AppContext } from "../contexts/AppContext";
import { useRequestState } from "../hooks/common";
import {
  AuthorizeGitlabResponse,
  GitProvider,
} from "../lib/api/applications/applications.pb";
import { gitlabOAuthRedirectURI } from "../lib/utils";

type Props = {
  className?: string;
  code: string;
  provider: GitProvider;
};

function OAuthCallback({ className, code, provider }: Props) {
  const history = useHistory();
  const {
    applicationsClient,
    storeProviderToken,
    getCallbackState,
    linkResolver,
  } = React.useContext(AppContext);
  const [res, loading, error, req] = useRequestState<AuthorizeGitlabResponse>();

  React.useEffect(() => {
    if (provider === GitProvider.GitLab) {
      const redirectUri = gitlabOAuthRedirectURI();

      req(
        applicationsClient.AuthorizeGitlab({
          redirectUri,
          code,
        })
      );
    }
  }, [code]);

  React.useEffect(() => {
    if (!res) {
      return;
    }

    storeProviderToken(GitProvider.GitLab, res.token);

    const state = getCallbackState();

    if (state?.page) {
      history.push(linkResolver(state.page));
      return;
    }
  }, [res]);

  return (
    <Page className={className}>
      <Flex wide align center>
        {loading && <CircularProgress />}
        {error && (
          <Alert
            title="Error completing OAuth 2.0 flow"
            severity="error"
            message={error.message}
          />
        )}
      </Flex>
    </Page>
  );
}

export default styled(OAuthCallback).attrs({ className: OAuthCallback.name })``;
