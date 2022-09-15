import { Dialog } from "@material-ui/core";
import qs from "query-string";
import * as React from "react";
import styled from "styled-components";
import { useFeatureFlags } from "../hooks/featureflags";
import { Alert, CrossNamespaceObjectRef, Kind } from "../lib/objects";
import { V2Routes } from "../lib/types";
import { statusSortHelper } from "../lib/utils";
import DataTable, {
  Field,
  filterByStatusCallback,
  filterConfig,
} from "./DataTable";
import { filterSeparator } from "./FilterDialog";
import KubeStatusIndicator from "./KubeStatusIndicator";
import Link from "./Link";
import Text from "./Text";
import YamlView from "./YamlView";
type Props = {
  className?: string;
  rows?: Alert[];
};

export const makeEventSourceLink = (obj: CrossNamespaceObjectRef) => {
  const url =
    obj.kind === Kind.Kustomization || obj.kind === Kind.HelmRelease
      ? V2Routes.Automations
      : V2Routes.Sources;
  let filters = `type${filterSeparator}${obj.kind}_`;
  if (obj.name !== "*") filters += `name${filterSeparator}${obj.name}_`;
  if (obj.namespace !== "*")
    filters += `namespace${filterSeparator}${obj.namespace}_`;
  return url + `?${qs.stringify({ filters: filters })}`;
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

  const [yamlView, setYamlView] = React.useState<Alert>(null);

  const alertFields: Field[] = [
    {
      label: "Name",
      value: (a) => (
        <Text
          onClick={() => setYamlView(a)}
          color="primary10"
          className="pointer"
        >
          {a.name}
        </Text>
      ),
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
            {a?.eventSources?.map((obj: CrossNamespaceObjectRef, index) => {
              return obj.name && obj.namespace && obj.kind ? (
                <Link
                  className="event-sources"
                  key={index}
                  to={makeEventSourceLink(obj)}
                >
                  {obj.kind}: {obj.namespace}/{obj.name}
                </Link>
              ) : (
                <li className="event-sources" key={index}>
                  {obj.kind}: {obj.namespace}/{obj.name}
                </li>
              );
            })}
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
    <>
      <DataTable
        className={className}
        fields={alertFields}
        rows={rows}
        filters={initialFilterState}
      />
      <Dialog open={yamlView !== null} onClose={() => setYamlView(null)}>
        {yamlView && (
          <YamlView
            object={yamlView}
            yaml={yamlView?.yaml}
            className="dialog"
          />
        )}
      </Dialog>
    </>
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
  ${Link}, li {
    &.event-sources {
      display: block;
    }
  }
`;
