import * as React from "react";
import styled from "styled-components";
import DataTable, { Field } from "../../components/DataTable";
import Page from "../../components/Page";
import { useFeatureFlags } from "../../hooks/featureflags";
import { useListProviders } from "../../hooks/notifications";

type Props = {
  className?: string;
};

function Settings({ className }: Props) {
  const { data: flagData } = useFeatureFlags();
  const flags = flagData?.flags || {};
  const { data, isLoading, error } = useListProviders();
  const providerFields: Field[] = [
    {
      label: "provider",
      value: (obj) => obj.name,
      textSearchable: true,
      defaultSort: true,
    },
    {
      label: "namespace",
      value: (obj) => obj.namespace,
    },
    ...(flags.WEAVE_GITOPS_FEATURE_CLUSTER === "true"
      ? [{ label: "Cluster", value: (obj) => obj.clusterName }]
      : []),
  ];

  return (
    <Page
      className={className}
      loading={isLoading}
      error={data?.errors || error}
    >
      <DataTable fields={providerFields} rows={data?.objects} />
    </Page>
  );
}

export default styled(Settings).attrs({ className: Settings.name })``;
