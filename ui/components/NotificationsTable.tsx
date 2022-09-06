import * as React from "react";
import styled from "styled-components";
import { useFeatureFlags } from "../hooks/featureflags";
import { Provider } from "../lib/objects";
import DataTable, { Field, filterConfig } from "./DataTable";

type Props = {
  className?: string;
  rows: Provider[];
};

function NotificationsTable({ className, rows }: Props) {
  const { data: flagData } = useFeatureFlags();
  const flags = flagData?.flags || {};

  let initialFilterState = {
    ...filterConfig(rows, "provider"),
    ...filterConfig(rows, "channel"),
    ...filterConfig(rows, "namespace"),
  };
  if (flags.WEAVE_GITOPS_FEATURE_CLUSTER === "true") {
    initialFilterState = {
      ...initialFilterState,
      ...filterConfig(rows, "clusterName"),
    };
  }
  if (flags.WEAVE_GITOPS_FEATURE_TENANCY === "true") {
    initialFilterState = {
      ...initialFilterState,
      ...filterConfig(rows, "tenant"),
    };
  }

  const providerFields: Field[] = [
    {
      label: "Name",
      value: "name",
      textSearchable: true,
      defaultSort: true,
    },
    {
      label: "Type",
      value: "provider",
    },
    {
      label: "Channel",
      value: "channel",
    },
    {
      label: "Namespace",
      value: "namespace",
    },
    ...(flags.WEAVE_GITOPS_FEATURE_CLUSTER === "true"
      ? [{ label: "Cluster", value: (obj) => obj.clusterName }]
      : []),
    ...(flags.WEAVE_GITOPS_FEATURE_TENANCY === "true"
      ? [{ label: "Tenant", value: "tenant" }]
      : []),
  ];

  return (
    <DataTable
      className={className}
      rows={rows}
      fields={providerFields}
      filters={initialFilterState}
    />
  );
}

export default styled(NotificationsTable).attrs({
  className: NotificationsTable.name,
})``;
