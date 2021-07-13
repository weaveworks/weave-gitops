import * as React from "react";
import styled from "styled-components";
import Page from "../components/Page";
import useAuth from "../hooks/auth";
import useNavigation from "../hooks/navigation";
import { PageRoute } from "../lib/types";

type Props = {
  className?: string;
};

function OauthCallback({ className }: Props) {
  const { authenticate, error } = useAuth();
  const { query, navigate } =
    useNavigation<{ code: string; provider: string; next?: string }>();

  React.useEffect(() => {
    authenticate(query.code, query.provider).then(() => {
      navigate(query.next ? `/${query.next}` : PageRoute.Applications);
    });
  }, [query.code]);

  return <Page error={error} className={className} loading></Page>;
}

export default styled(OauthCallback)``;
