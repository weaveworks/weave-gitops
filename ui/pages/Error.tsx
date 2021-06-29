import * as React from "react";
import styled from "styled-components";
import Flex from "../components/Flex";
import Page from "../components/Page";

type Props = {
  className?: string;
  code?: string;
};

function ErrorPage({ className }: Props) {
  return (
    <Page className={className} title="Error">
      <Flex wide center>
        <h2>404</h2>
      </Flex>
    </Page>
  );
}

export default styled(ErrorPage)``;
