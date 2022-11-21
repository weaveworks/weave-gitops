import * as React from "react";
import styled from "styled-components";
import DataTable, { filterConfig } from "./DataTable";

export interface FluxVersion {
  version?: string;
  clusterName?: string;
  namespace?: string;
}
type Props = {
  className?: string;
  versions?: FluxVersion[];
};

function FluxVersionsTable({ className, versions = [] }: Props) {
  const initialFilterState = {
    ...filterConfig(versions, "version"),
  };

  return (
    <DataTable
      className={className}
      filters={initialFilterState}
      rows={versions}
      fields={[
        {
          label: "Cluster",
          value: "clusterName",
          textSearchable: true,
          maxWidth: 600,
        },
        {
          label: "Namespace",
          value: "namespace",
        },
        {
          label: "Flux Version",
          value: "version",
        },
      ]}
    />
  );
}

export default styled(FluxVersionsTable).attrs({
  className: FluxVersionsTable.name,
})``;
