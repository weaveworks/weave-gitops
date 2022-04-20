import * as React from "react";
import styled from "styled-components";
import Page from "../../components/Page";
import HelmRepositoryDetailComponent from "../../components/HelmRepositoryDetail";

type Props = {
  className?: string;
  name: string;
  namespace: string;
};

function HelmRepositoryDetail({ className, name, namespace }: Props) {
  return (
    <Page error={null} className={className} title={name}>
      <HelmRepositoryDetailComponent
        name={name}
        namespace={namespace}
      />
    </Page>
  );
}

export default styled(HelmRepositoryDetail).attrs({
  className: HelmRepositoryDetail.name,
})``;
