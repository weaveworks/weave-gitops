import * as React from "react";
import styled from "styled-components";
import OCIRepositoryDetail from "../../components/OCIRepositoryDetail";
import Page from "../../components/Page";

type Props = {
  className?: string;
  name: string;
  namespace: string;
  clusterName: string;
};

function OCIRepositoryPage({ className, name, namespace, clusterName }: Props) {
  return (
    <Page error={null} className={className}>
      <OCIRepositoryDetail
        name={name}
        namespace={namespace}
        clusterName={clusterName}
      />
    </Page>
  );
}

export default styled(OCIRepositoryPage).attrs({
  className: OCIRepositoryPage.name,
})``;
