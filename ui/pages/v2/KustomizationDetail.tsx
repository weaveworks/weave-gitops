import * as React from "react";
import styled from "styled-components";
import KustomizationComponent from "../../components/KustomizationDetail";
import Page from "../../components/Page";
import { useGetKustomization } from "../../hooks/automations";

type Props = {
  name: string;
  clusterName: string;
  className?: string;
};
function KustomizationDetail({ className, name, clusterName }: Props) {
  const { data, isLoading, error } = useGetKustomization(name, clusterName);
  const kustomization = data?.kustomization;
  return (
    <Page loading={isLoading} error={error} className={className} title={name}>
      <KustomizationComponent kustomization={kustomization} name={name} />
    </Page>
  );
}

export default styled(KustomizationDetail).attrs({
  className: KustomizationDetail.name,
})``;
