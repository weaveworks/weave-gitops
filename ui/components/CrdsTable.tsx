import * as React from "react";
import styled from "styled-components";
import { Crd } from "../lib/api/core/types.pb";
import { filterConfig } from "./FilterableTable";
import URLAddressableTable from "./URLAddressableTable";

type Props = {
  className?: string;
  crds?: Crd[];
};

function CrdsTable({ className, crds = [] }: Props) {
  const initialFilterState = {
    ...filterConfig(crds, "version"),
    ...filterConfig(crds, "kind"),
    ...filterConfig(crds, "clusterName"),
  };
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
        {
          label: "Cluster",
          value: "clusterName",
        },
      ]}
    />
  );
}

export default styled(CrdsTable).attrs({ className: CrdsTable.name })``;
