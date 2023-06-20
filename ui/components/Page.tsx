import _ from "lodash";
import React from "react";
import styled from "styled-components";
import useCommon from "../hooks/common";
import { MultiRequestError, RequestError } from "../lib/types";
import Alert from "./Alert";
import Breadcrumbs, { Breadcrumb } from "./Breadcrumbs";
import Flex from "./Flex";
import LoadingPage from "./LoadingPage";
import Spacer from "./Spacer";
import UserSettings from "./UserSettings";

export type PageProps = {
  className?: string;
  children?: any;
  loading?: boolean;
  path: Breadcrumb[];
  error?: RequestError | RequestError[] | MultiRequestError[];
};

export const topBarHeight = "60px";

const ContentContainer = styled.div`
  height: 100%;
  width: calc(100% - 48px);
  padding: 0 ${(props) => props.theme.spacing.medium};
  max-height: calc(100vh - ${topBarHeight});
  overflow-wrap: normal;
  overflow-x: scroll;
  margin: 0px auto;
`;
const PageLayout = styled(Flex)`
  width: 100%;
  flex-grow: 1;
  overflow: hidden;
`;
export const Content = styled(Flex)`
  background-color: ${(props) => props.theme.colors.white};
  border-radius: 10px;
  box-sizing: border-box;
  margin: 0 auto;
  min-height: 100%;
`;

const Children = styled(Flex)`
  width: calc(100% - 48px);
  padding: ${(props) => props.theme.spacing.medium};
  height: 100%;
`;

const TopToolBar = styled(Flex)`
  height: ${topBarHeight};
  min-width: 650px;
  z-index: 2;
  width: calc(100% - 64px);
  padding: 0 ${(props) => props.theme.spacing.large};
`;

export function Errors({ error }) {
  const arr = _.isArray(error) ? error : [error];
  if (arr[0])
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
  return null;
}

function Page({ children, loading, error, className, path }: PageProps) {
  const { settings } = useCommon();
  return (
    <PageLayout column wide tall>
      <TopToolBar start align wide between>
        <Breadcrumbs path={path} />
        <UserSettings />
      </TopToolBar>
      <ContentContainer>
        <Content wide between column className={className}>
          {loading ? (
            <LoadingPage />
          ) : (
            <>
              <Children column wide tall start>
                <Errors error={error} />
                {children}
              </Children>
              {settings.footer}
            </>
          )}
        </Content>
      </ContentContainer>
    </PageLayout>
  );
}

export default styled(Page)`
  .MuiAlert-root {
    width: 100%;
  }
`;
