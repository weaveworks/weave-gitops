import _ from "lodash";
import React from "react";
import { useIsFetching } from "react-query";
import styled from "styled-components";
import useCommon from "../hooks/common";
import { PageRoute, RequestError } from "../lib/types";
import Alert from "./Alert";
import Flex from "./Flex";
import Footer from "./Footer";
import LoadingPage from "./LoadingPage";
import PollingIndicator from "./PollingIndicator";
import Spacer from "./Spacer";
import Text from "./Text";

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

export const Content = styled(Flex)`
  background-color: rgba(255, 255, 255, 0.75);
  border-radius: 10px;
  box-sizing: border-box;
  margin: 0 auto;
  min-height: 480px;
  max-width: 100%;
  padding-bottom: ${(props) => props.theme.spacing.medium};
  padding-left: ${(props) => props.theme.spacing.medium};
  padding-right: ${(props) => props.theme.spacing.medium};
  padding-top: ${(props) => props.theme.spacing.medium};
  overflow: hidden;
`;

const Children = styled(Flex)``;

export const TitleBar = styled(Flex)`
  //matches nav tabs
  line-height: 1.75;
`;

function Errors({ error }) {
  const arr = _.isArray(error) ? error : [error];
  return (
    <Flex wide column>
      <Spacer padding="xs" />
      {_.map(arr, (e, i) => (
        <Flex key={i} wide start>
          <Alert title="Error" message={e?.message} severity="error" />
        </Flex>
      ))}
      <Spacer padding="xs" />
    </Flex>
  );
}

function Page({
  children,
  title,
  actions,
  loading,
  error,
  className,
}: PageProps) {
  const { settings } = useCommon();
  const fetching = useIsFetching();

  if (loading) {
    return (
      <Content wide tall start column>
        <LoadingPage />
      </Content>
    );
  }

  return (
    <Content wide tall start column className={className}>
      <TitleBar wide start between>
        <Flex align>
          <Text semiBold size="large">
            {title}
          </Text>
          <Spacer padding="small" />
          <PollingIndicator loading={fetching > 0} />
        </Flex>
        {actions}
      </TitleBar>
      {error && <Errors error={error} />}
      <Children column wide tall start>
        {children}
      </Children>

      {settings.renderFooter && <Footer />}
    </Content>
  );
}

export default styled(Page)`
  .MuiAlert-root {
    width: 100%;
  }
`;
