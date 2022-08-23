import * as React from "react";
import styled from "styled-components";
import { Crd } from "../lib/api/core/types.pb";
import { useFeatureFlags } from "../hooks/featureflags";
import { filterConfig } from "./FilterableTable";
import URLAddressableTable from "./URLAddressableTable";

type Props = {
  className?: string;
  crds?: Crd[];
};

function CrdsTable({ className, crds = [] }: Props) {
  const { data } = useFeatureFlags();
  const flags = data?.flags || {};

  let initialFilterState = {
    ...filterConfig(crds, "version"),
    ...filterConfig(crds, "kind"),
  };

  if (flags.WEAVE_GITOPS_FEATURE_CLUSTER === "true") {
    initialFilterState = {
      ...initialFilterState,
      ...filterConfig(crds, "clusterName"),
    };
  }

  return (
    <URLAddressableTable
      className={className}
      filters={initialFilterState}
      rows={crds}
      fields={[
        {
          label: "Name",
          value: (d: Crd) => d.name.plural + "." + d.name.group,
          textSearchable: true,
          maxWidth: 600,
        },
        {
          label: "Kind",
          value: "kind",
        },
        {
          label: "Version",
          value: "version",
        },
        ...(flags.WEAVE_GITOPS_FEATURE_CLUSTER === "true"
          ? [{ label: "Cluster", value: "clusterName" }]
          : []),
      ]}
    />
  );
}

export default styled(CrdsTable).attrs({ className: CrdsTable.name })``;
