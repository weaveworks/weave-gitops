import * as React from "react";
import styled from "styled-components";
import HelmRepositoryDetailComponent from "../../components/HelmRepositoryDetail";
import Page from "../../components/Page";
import { useGetObject } from "../../hooks/objects";
import { HelmRepository, Kind } from "../../lib/objects";

type Props = {
  className?: string;
  name: string;
  namespace: string;
  clusterName: string;
};

function HelmRepositoryDetail({
  className,
  name,
  namespace,
  clusterName,
}: Props) {
  const {
    data: helmRepository,
    isLoading,
    error,
  } = useGetObject<HelmRepository>(
    name,
    namespace,
    Kind.HelmRepository,
    clusterName
  );

  return (
    <Page error={error} loading={isLoading} className={className} title={name}>
      <HelmRepositoryDetailComponent helmRepository={helmRepository} />
    </Page>
  );
}

export default styled(HelmRepositoryDetail).attrs({
  className: HelmRepositoryDetail.name,
})``;
