import * as React from "react";
import styled from "styled-components";
import Content from "../components/Page";
import Flex from "../components/Flex";
import { FeatureFlags } from "../contexts/FeatureFlags";
import LoadingPage from "../components/LoadingPage";

type Props = {
  className?: string;
  code?: string;
};

function ErrorPage({ className }: Props) {
  const { loading } = React.useContext(FeatureFlags);

  const Error404Animation = React.lazy(
    () => import("../components/Animations/Error404")
  );

  return (
    <Content>
      {loading ? (
        <LoadingPage />
      ) : (
        <Flex center wide align>
          <React.Suspense fallback={null}>
            <Error404Animation />
          </React.Suspense>
        </Flex>
      )}
    </Content>
  );
}

export default styled(ErrorPage)``;
