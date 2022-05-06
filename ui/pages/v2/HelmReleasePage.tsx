import * as React from "react";
import styled from "styled-components";
import HelmReleaseDetail from "../../components/HelmReleaseDetail";
import Page from "../../components/Page";
import { useGetHelmRelease } from "../../hooks/automations";

type Props = {
  name: string;
  namespace: string;
  clusterName: string;
  className?: string;
};

function HelmReleasePage({ className, name, namespace, clusterName }: Props) {
  const { data, isLoading, error } = useGetHelmRelease(
    name,
    namespace,
    clusterName
  );
  const helmRelease = data?.helmRelease;

  return (
    <Page loading={isLoading} error={error} className={className} title={name}>
      <HelmReleaseDetail
        helmRelease={helmRelease}
        name={name}
        clusterName={clusterName}
      />
    </Page>
  );
}

export default styled(HelmReleasePage).attrs({
  className: HelmReleasePage.name,
})``;
