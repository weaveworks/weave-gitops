import * as React from "react";
import styled from "styled-components";
import HelmReleaseComponent from "../../components/HelmReleaseDetail";
import Page from "../../components/Page";
import { useGetHelmRelease } from "../../hooks/automations";

type Props = {
  name: string;
  clusterName: string;
  className?: string;
};

function HelmReleaseDetail({ className, name, clusterName }: Props) {
  const { data, isLoading, error } = useGetHelmRelease(name, clusterName);
  const helmRelease = data?.helmRelease;

  return (
    <Page loading={isLoading} error={error} className={className} title={name}>
      <HelmReleaseComponent helmRelease={helmRelease} name={name} clusterName={clusterName} />
    </Page>
  );
}

export default styled(HelmReleaseDetail).attrs({
  className: HelmReleaseDetail.name,
})``;
