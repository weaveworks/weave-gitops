import * as React from "react";
import styled from "styled-components";
import Page from "../../components/Page";
import GitRepositoryDetailComponent from "../../components/GitRepositoryDetail";

type Props = {
  className?: string;
  name: string;
  namespace: string;
};

function GitRepositoryDetail({ className, name, namespace }: Props) {
  return (
    <Page error={null} className={className} title={name}>
      <GitRepositoryDetailComponent
        name={name}
        namespace={namespace}
      />
    </Page>
  );
}

export default styled(GitRepositoryDetail).attrs({
  className: GitRepositoryDetail.name,
})``;
