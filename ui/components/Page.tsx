import { Breadcrumbs } from "@material-ui/core";
import _ from "lodash";
import React from "react";
import styled from "styled-components";
import useCommon from "../hooks/common";
import { PageRoute } from "../lib/types";
import { formatURL } from "../lib/utils";
import Alert from "./Alert";
import Flex from "./Flex";
import Link from "./Link";
import LoadingPage from "./LoadingPage";

export type PageProps = {
  className?: string;
  children?: any;
  title?: string;
  breadcrumbs?: { page: PageRoute; query?: any }[];
  loading?: boolean;
};

const Content = styled.div`
  max-width: 1400px;
  margin: 0 auto;
  width: 100%;
  background-color: ${(props) => props.theme.colors.white};
  padding-left: ${(props) => props.theme.spacing.large};
  padding-right: ${(props) => props.theme.spacing.large};
  padding-top: ${(props) => props.theme.spacing.large};
  padding-bottom: ${(props) => props.theme.spacing.medium};
`;

export const TitleBar = styled.div`
  display: flex;
  align-items: center;
  margin-bottom: ${(props) => props.theme.spacing.small};

  h2 {
    margin: 0 !important;
  }
`;

function pageLookup(p: PageRoute) {
  switch (p) {
    case PageRoute.Applications:
      return "Applications";

    default:
      break;
  }
}

function Page({ className, children, title, breadcrumbs, loading }: PageProps) {
  const { appState } = useCommon();

  if (loading) {
    return (
      <Content>
        <LoadingPage />
      </Content>
    );
  }

  return (
    <div className={className}>
      <Content>
        <TitleBar>
          <Breadcrumbs>
            {breadcrumbs &&
              _.map(breadcrumbs, (b) => (
                <Link key={b.page} to={formatURL(b.page, b.query)}>
                  <h2>{pageLookup(b.page)}</h2>
                </Link>
              ))}
            <h2>{title}</h2>
          </Breadcrumbs>
        </TitleBar>
        {appState.error && (
          <Flex center wide>
            <Alert
              title={appState.error.message}
              message={appState.error.detail}
              severity="error"
            />
          </Flex>
        )}
        <div>{children}</div>
      </Content>
    </div>
  );
}

export default styled(Page)`
  display: flex;

  .MuiAlert-root {
    width: 100%;
  }
`;
