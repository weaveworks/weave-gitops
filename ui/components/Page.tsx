import _ from "lodash";
import React from "react";
import styled from "styled-components";
import useCommon from "../hooks/common";
import { MultiRequestError, RequestError } from "../lib/types";
import Alert from "./Alert";
import Flex from "./Flex";
import Footer from "./Footer";
import LoadingPage from "./LoadingPage";
import Spacer from "./Spacer";
import UserSettings from "./UserSettings";
import Breadcrumbs from "./Breadcrumbs";

export type PageProps = {
  className?: string;
  children?: any;
  loading?: boolean;
  error?: RequestError | RequestError[] | MultiRequestError[];
};
const ContentContainer = styled.div`
  width: 100%;
  display: flex;
  overflow: hidden;
  flex-grow: 1;
  
`;
const PageLayout = styled(Flex)`
  padding: 0 ${(props) => props.theme.spacing.medium};
`;
export const Content = styled(Flex)`
  background-color: ${(props) => props.theme.colors.white};
  border-radius: 10px;
  box-sizing: border-box;
  margin: 0 auto;
  min-height: 100%;
  max-width: 100%;
`;
const topBarHeight = "60px";

const Children = styled(Flex)`
  max-height: calc(100vh - ${topBarHeight});
  overflow-wrap: normal;
  overflow-x: scroll;
  padding-right: 24px;
  margin: 0px auto;
`;

const TopToolBar = styled(Flex)`
  height: ${topBarHeight};
  min-width: 650px;
  z-index: 2;
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

function Page({ children, loading, error, className }: PageProps) {
  const { settings } = useCommon();
  return (
    <PageLayout column wide tall>
      <TopToolBar start align wide between>
        <Breadcrumbs />
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
                {/* {children} */}
              </Children>
              {settings.renderFooter && <Footer />}
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
