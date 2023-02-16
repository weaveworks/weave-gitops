import * as React from "react";
import styled from "styled-components";
import KustomizationDetail from "../../components/KustomizationDetail";
import Page from "../../components/Page";
import { useGetObject } from "../../hooks/objects";
import { Kind } from "../../lib/api/core/types.pb";
import { Kustomization } from "../../lib/objects";

type Props = {
  name: string;
  namespace?: string;
  clusterName: string;
  className?: string;
};

function KustomizationPage({ className, name, namespace, clusterName }: Props) {
  console.log("KustomizationPage", name, namespace, clusterName);
  const {
    data: kustomization,
    isLoading,
    error,
  } = useGetObject<Kustomization>(
    name,
    namespace,
    Kind.Kustomization,
    clusterName
  );
  return (
    <Page loading={isLoading} error={error} className={className}>
      <KustomizationDetail kustomization={kustomization} />
    </Page>
  );
}

export default styled(KustomizationPage).attrs({
  className: KustomizationPage.name,
})``;
