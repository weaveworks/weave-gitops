import * as React from "react";
import styled from "styled-components";
import GitRepositoryDetailComponent from "../../components/GitRepositoryDetail";
import Page from "../../components/Page";

type Props = {
  className?: string;
  name: string;
  namespace: string;
};

function GitRepositoryDetail({ className, name, namespace }: Props) {
  return (
    <Page error={null} className={className}>
      <GitRepositoryDetailComponent name={name} namespace={namespace} />
    </Page>
  );
}

export default styled(GitRepositoryDetail).attrs({
  className: GitRepositoryDetail.name,
})``;
