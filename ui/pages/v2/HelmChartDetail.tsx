import * as React from "react";
import styled from "styled-components";
import Page from "../../components/Page";
import HelmChartDetailComponent from "../../components/HelmChartDetail";

type Props = {
  className?: string;
  name: string;
  namespace: string;
};

function HelmChartDetail({ className, name, namespace }: Props) {
  return (
    <Page error={null} className={className} title={name}>
      <HelmChartDetailComponent
        name={name}
        namespace={namespace}
      />
    </Page>
  );
}

export default styled(HelmChartDetail).attrs({
  className: HelmChartDetail.name,
})``;
