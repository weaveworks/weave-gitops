import * as React from "react";
import styled from "styled-components";
import HelmRepositoryDetailComponent from "../../components/HelmRepositoryDetail";
import Page from "../../components/Page";

type Props = {
  className?: string;
  name: string;
  namespace: string;
};

function HelmRepositoryDetail({ className, name, namespace }: Props) {
  return (
    <Page error={null} className={className}>
      <HelmRepositoryDetailComponent name={name} namespace={namespace} />
    </Page>
  );
}

export default styled(HelmRepositoryDetail).attrs({
  className: HelmRepositoryDetail.name,
})``;
