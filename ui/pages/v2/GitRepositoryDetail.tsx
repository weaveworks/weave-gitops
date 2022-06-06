import * as React from "react";
import styled from "styled-components";
import GitRepositoryDetailComponent from "../../components/GitRepositoryDetail";
import Page from "../../components/Page";

type Props = {
  className?: string;
  name: string;
  namespace: string;
  clusterName: string;
};

function GitRepositoryDetail({ className, name, namespace, clusterName }: Props) {
  return (
    <Page error={null} className={className}>
      <GitRepositoryDetailComponent name={name} namespace={namespace} clusterName={clusterName} />
    </Page>
  );
}

export default styled(GitRepositoryDetail).attrs({
  className: GitRepositoryDetail.name,
})``;
