import * as React from "react";
import styled from "styled-components";
import Alert from "./Alert";
import Page, { PageProps } from "./Page";

type Props = {
  className?: string;
  error: { title: string; message: string };
} & PageProps;

function ErrorPage({ className, error, ...rest }: Props) {
  return (
    <Page {...rest} className={className} loading={false}>
      <Alert severity="error" title={error.title} message={error.message} />
    </Page>
  );
}

export default styled(ErrorPage).attrs({ className: ErrorPage.name })``;
