import * as React from "react";
import styled from "styled-components";
import HelmRepositoryDetailComponent from "../../components/HelmRepositoryDetail";
import Page from "../../components/Page";

type Props = {
  className?: string;
  name: string;
  namespace: string;
  clusterName: string;
};

function HelmRepositoryDetail({ className, name, namespace, clusterName }: Props) {
  return (
    <Page error={null} className={className}>
      <HelmRepositoryDetailComponent name={name} namespace={namespace} clusterName={clusterName} />
    </Page>
  );
}

export default styled(HelmRepositoryDetail).attrs({
  className: HelmRepositoryDetail.name,
})``;
