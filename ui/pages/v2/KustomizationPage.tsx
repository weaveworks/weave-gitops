import * as React from "react";
import styled from "styled-components";
import KustomizationDetail from "../../components/KustomizationDetail";
import Page from "../../components/Page";
import { useGetObject } from "../../hooks/objects";
import { Kind } from "../../lib/api/core/types.pb";
import { Kustomization } from "../../lib/objects";
import { V2Routes } from "../../lib/types";

type Props = {
  name: string;
  namespace?: string;
  clusterName: string;
  className?: string;
};

function KustomizationPage({ className, name, namespace, clusterName }: Props) {
  const {
    data: kustomization,
    isLoading,
    error,
  } = useGetObject<Kustomization>(
    name,
    namespace,
    Kind.Kustomization,
    clusterName,
  );
  return (
    <Page
      loading={isLoading}
      error={error}
      className={className}
      path={[
        { label: "Applications", url: V2Routes.Automations },
        { label: name },
      ]}
    >
      <KustomizationDetail kustomization={kustomization} />
    </Page>
  );
}

export default styled(KustomizationPage).attrs({
  className: KustomizationPage.name,
})``;
