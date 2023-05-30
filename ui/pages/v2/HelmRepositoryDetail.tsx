import * as React from "react";
import styled from "styled-components";
import HelmRepositoryDetailComponent from "../../components/HelmRepositoryDetail";
import Page from "../../components/Page";
import { useGetObject } from "../../hooks/objects";
import { Kind } from "../../lib/api/core/types.pb";
import { HelmRepository } from "../../lib/objects";
import { V2Routes } from "../../lib/types";

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
    <Page
      error={error}
      loading={isLoading}
      className={className}
      path={[{ label: "Sources", url: V2Routes.Sources }, { label: name }]}
    >
      <HelmRepositoryDetailComponent helmRepository={helmRepository} />
    </Page>
  );
}

export default styled(HelmRepositoryDetail).attrs({
  className: HelmRepositoryDetail.name,
})``;
