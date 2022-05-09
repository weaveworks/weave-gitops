import * as React from "react";
import styled from "styled-components";
import HelmChartDetailComponent from "../../components/HelmChartDetail";
import Page from "../../components/Page";

type Props = {
  className?: string;
  name: string;
  namespace: string;
};

function HelmChartDetail({ className, name, namespace }: Props) {
  return (
    <Page error={null} className={className}>
      <HelmChartDetailComponent name={name} namespace={namespace} />
    </Page>
  );
}

export default styled(HelmChartDetail).attrs({
  className: HelmChartDetail.name,
})``;
