import _ from "lodash";
import React from "react";
import styled from "styled-components";
import useCommon from "../hooks/common";
import { PageRoute, RequestError } from "../lib/types";
import Alert from "./Alert";
import Flex from "./Flex";
import Footer from "./Footer";
import LoadingPage from "./LoadingPage";
import PollingIndicator from "./PollingIndicator";

export type PageProps = {
  className?: string;
  children?: any;
  title?: string | JSX.Element;
  breadcrumbs?: { page: PageRoute; query?: any }[];
  actions?: JSX.Element;
  loading?: boolean;
  error?: RequestError | RequestError[];
  isFetching?: boolean;
};

export const Content = styled.div`
  background-color: rgba(255, 255, 255, 0.75);
  border-radius: 10px;
  box-sizing: border-box;
  margin: 0 auto;
  min-width: 1260px;
  min-height: 480px;
  padding-bottom: ${(props) => props.theme.spacing.medium};
  padding-left: ${(props) => props.theme.spacing.large};
  padding-top: ${(props) => props.theme.spacing.medium};
  width: 100%;
`;

const Children = styled.div``;

export const TitleBar = styled.div`
  align-items: center;
  display: flex;
  justify-content: space-between;
  width: 100%;
  h2 {
    margin: 0 !important;
    color: ${(props) => props.theme.colors.neutral40} !important;
  }
`;

function Errors({ error }) {
  const arr = _.isArray(error) ? error : [error];
  return (
    <>
      {_.map(arr, (e, i) => (
        <Flex key={i} center wide>
          <Alert title="Error" message={e?.message} severity="error" />
        </Flex>
      ))}
    </>
  );
}

function Page({
  className,
  children,
  title,
  actions,
  loading,
  error,
  isFetching,
}: PageProps) {
  const { settings } = useCommon();

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
          <Flex align>
            <h2>{title}</h2>
            {title && <PollingIndicator loading={isFetching} />}
          </Flex>
          {actions}
        </TitleBar>
        {error && <Errors error={error} />}
        <Children>{children}</Children>
      </Content>
      {settings.renderFooter && <Footer />}
    </div>
  );
}

export default styled(Page)`
  .MuiAlert-root {
    width: 100%;
  }
`;
