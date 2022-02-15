import _ from "lodash";
import * as React from "react";
import styled from "styled-components";
import { Button } from "../..";
import KustomizationTable from "../../components/KustomizationTable";
import Link from "../../components/Link";
import Page from "../../components/Page";
import { useGetRemoteKustomizations } from "../../hooks/kustomizations";
import { V2Routes } from "../../lib/types";
import { formatURL } from "../../lib/utils";

type Props = {
  className?: string;
};

function KustomizationList({ className }: Props) {
  // const { data: automations, error, isLoading } = useGetKustomizations();
  const { data, error, isLoading } = useGetRemoteKustomizations([
    "management-cluster",
    "leaf-cluster-1",
  ]);

  console.log(data);
  return (
    <Page
      error={error}
      loading={isLoading}
      title="Kustomizations"
      actions={
        <Link to={formatURL(V2Routes.AddKustomization)}>
          <Button>Add Kustomization</Button>
        </Link>
      }
      className={className}
    >
      <KustomizationTable
        kustomizations={_.map(data?.kustomizations, (k) => ({
          ...k.kustomization,
          cluster: k.clusterName,
        }))}
      />
    </Page>
  );
}

export default styled(KustomizationList).attrs({
  className: KustomizationList.name,
})``;
