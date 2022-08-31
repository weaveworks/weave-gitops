import * as React from "react";
import styled from "styled-components";
import HelmChartDetailComponent from "../../components/HelmChartDetail";
import Page from "../../components/Page";
import { useGetObject } from "../../hooks/objects";
import { HelmChart, Kind } from "../../lib/objects";

type Props = {
  className?: string;
  name: string;
  namespace: string;
  clusterName: string;
};

function HelmChartDetail({ className, name, namespace, clusterName }: Props) {
  const {
    data: helmChart,
    isLoading,
    error,
  } = useGetObject<HelmChart>(name, namespace, Kind.HelmChart, clusterName);

  return (
    <Page error={error} loading={isLoading} className={className} title={name}>
      <HelmChartDetailComponent helmChart={helmChart} />
    </Page>
  );
}

export default styled(HelmChartDetail).attrs({
  className: HelmChartDetail.name,
})``;
