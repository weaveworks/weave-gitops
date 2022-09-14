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
  rows?: Alert[];
};

function AlertsTable({ className, rows = [] }: Props) {
  const { data: flagData } = useFeatureFlags();
  const flags = flagData?.flags || {};
  let initialFilterState = {
    ...filterConfig(rows, "name"),
    ...filterConfig(rows, "namespace"),
    ...filterConfig(rows, "severity"),
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
      textSearchable: true,
    },
    {
      label: "Namespace",
      value: "namespace",
    },
    {
      label: "Severity",
      value: "severity",
    },
    {
      label: "Event Sources",
      value: (a) => {
        return (
          <ul className="event-sources">
            {a?.eventSources?.map((obj: CrossNamespaceObjectRef) => (
              <li className="event-sources" key={obj.name}>
                {obj.kind}: {obj.namespace}/{obj.name}
              </li>
            ))}
          </ul>
        );
      },
      labelRenderer: () => <h2 className="event-sources">Event Sources</h2>,
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
  //these styles did not apply when wrapped in .event-sources, only in this more repetitive format
  ul {
    &.event-sources {
      padding: 0px;
    }
  }
  h2 {
    &.event-sources {
      padding-left: ${(props) => props.theme.spacing.small};
    }
  }
  li {
    &.event-sources {
      display: block;
    }
  }
`;
