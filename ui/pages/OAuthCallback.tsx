import * as React from "react";
import styled from "styled-components";
import Page from "../components/Page";
import useAuth from "../hooks/auth";
import useNavigation from "../hooks/navigation";
import { OauthProviderName } from "../lib/api/applications/applications.pb";
import { PageRoute } from "../lib/types";

type Props = {
  className?: string;
};

function OauthCallback({ className }: Props) {
  const { authenticate } = useAuth();
  const { query, navigate } =
    useNavigation<{ code: string; provider: string }>();

  React.useEffect(() => {
    authenticate(query.code, query.provider as OauthProviderName).then(() =>
      navigate(PageRoute.Applications)
    );
  }, [query.code]);

  return <Page className={className} loading></Page>;
}

export default styled(OauthCallback)``;
