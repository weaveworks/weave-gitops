import * as React from "react";
import styled from "styled-components";
import HelmReleaseDetail from "../../components/HelmReleaseDetail";
import Page from "../../components/Page";
import { useGetObject } from "../../hooks/objects";
import { HelmRelease, Kind } from "../../lib/objects";

type Props = {
  name: string;
  namespace: string;
  clusterName: string;
  className?: string;
};

function HelmReleasePage({ className, name, namespace, clusterName }: Props) {
  const {
    data: helmRelease,
    isLoading,
    error,
  } = useGetObject<HelmRelease>(name, namespace, Kind.HelmRelease, clusterName);

  return (
    <Page loading={isLoading} error={error} className={className}>
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
