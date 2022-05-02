import * as React from "react";
import styled from "styled-components";
import KustomizationComponent from "../../components/KustomizationDetail";
import Page from "../../components/Page";
import { useGetKustomization } from "../../hooks/automations";

type Props = {
  name: string;
  namespace?: string;
  clusterName: string;
  className?: string;
};

function KustomizationPage({ className, name, namespace, clusterName }: Props) {
  const { data, isLoading, error } = useGetKustomization(
    name,
    namespace,
    clusterName
  );
  const kustomization = data?.kustomization;
  return (
    <Page loading={isLoading} error={error} className={className}>
      <KustomizationComponent kustomization={kustomization} />
    </Page>
  );
}

export default styled(KustomizationPage).attrs({
  className: KustomizationPage.name,
})``;
