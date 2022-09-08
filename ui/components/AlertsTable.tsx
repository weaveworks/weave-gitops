import * as React from "react";
import styled from "styled-components";
import { useFeatureFlags } from "../hooks/featureflags";
import { Alert, CrossNamespaceObjectRef } from "../lib/objects";
import { statusSortHelper } from "../lib/utils";
import DataTable, {
  Field,
  filterByStatusCallback,
  filterConfig,
} from "./DataTable";
import KubeStatusIndicator from "./KubeStatusIndicator";
type Props = {
  className?: string;
  rows?: any[];
};

function AlertsTable({ className, rows = [] }: Props) {
  const { data: flagData } = useFeatureFlags();
  const flags = flagData?.flags || {};
  let initialFilterState = {
    ...filterConfig(rows, "name"),
    ...filterConfig(rows, "namespace"),
    ...filterConfig(rows, "eventSeverity"),
    ...filterConfig(rows, "status", filterByStatusCallback),
  };
  if (flags.WEAVE_GITOPS_FEATURE_CLUSTER === "true") {
    initialFilterState = {
      ...initialFilterState,
      ...filterConfig(rows, "clusterName"),
    };
  }

  const alertFields: Field[] = [
    {
      label: "Name",
      value: "name",
    },
    {
      label: "Namespace",
      value: "namespace",
    },
    {
      label: "Severity",
      value: "eventSeverity",
    },
    {
      label: "Event Sources",
      value: (a) => {
        return (
          <ul>
            {a?.eventSources?.map((obj: CrossNamespaceObjectRef) => (
              <li key={obj.name}>
                {obj.kind}: {obj.name}
              </li>
            ))}
          </ul>
        );
      },
      labelRenderer: () => (
        <h2 style={{ paddingLeft: "12px" }}>Event Sources</h2>
      ),
    },
    {
      label: "Status",
      value: (a: Alert) =>
        a.conditions.length > 0 ? (
          <KubeStatusIndicator
            short
            conditions={a.conditions}
            suspended={a.suspended}
          />
        ) : null,
      sortValue: statusSortHelper,
    },
    ...(flags.WEAVE_GITOPS_FEATURE_CLUSTER === "true"
      ? [{ label: "Cluster", value: (obj) => obj.clusterName }]
      : []),
  ];

  return (
    <DataTable
      className={className}
      fields={alertFields}
      rows={rows}
      filters={initialFilterState}
    />
  );
}

export default styled(AlertsTable).attrs({ className: AlertsTable.name })`
  ul {
    padding: 0px;
  }
  li {
    display: block;
  }
`;
