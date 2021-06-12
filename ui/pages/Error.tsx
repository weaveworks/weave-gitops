import * as React from "react";
import styled from "styled-components";
import Page from "../components/Page";

type Props = {
  className?: string;
  code?: string;
};

function ErrorPage({ className }: Props) {
  return (
    <Page className={className} title="Error">
      <h2>404</h2>
    </Page>
  );
}

export default styled(ErrorPage)``;
