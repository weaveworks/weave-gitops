import _ from "lodash";
import * as React from "react";
import styled from "styled-components";
import { Automation } from "../hooks/automations";
import { useFeatureFlags } from "../hooks/featureflags";
import { FluxObjectKind, HelmRelease } from "../lib/api/core/types.pb";
import { formatURL } from "../lib/nav";
import { V2Routes } from "../lib/types";
import {
  removeKind,
  statusSortHelper,
  getSourceRefForAutomation,
} from "../lib/utils";
import DataTable, {
  Field,
  filterByStatusCallback,
  filterConfig,
} from "./DataTable";
import KubeStatusIndicator, { computeMessage } from "./KubeStatusIndicator";
import Link from "./Link";
import SourceLink from "./SourceLink";
import Timestamp from "./Timestamp";

type Props = {
  className?: string;
  automations?: Automation[];
  appName?: string;
  hideSource?: boolean;
};

function AutomationsTable({ className, automations, hideSource }: Props) {
  const { data } = useFeatureFlags();
  const flags = data?.flags || {};

  automations = automations.map((a) => {
    return { ...a, type: removeKind(a.kind) };
  });

  let initialFilterState = {
    ...filterConfig(automations, "type"),
    ...filterConfig(automations, "namespace"),
    ...filterConfig(automations, "status", filterByStatusCallback),
  };

  if (flags.WEAVE_GITOPS_FEATURE_TENANCY === "true") {
    initialFilterState = {
      ...initialFilterState,
      ...filterConfig(automations, "tenant"),
    };
  }

  if (flags.WEAVE_GITOPS_FEATURE_CLUSTER === "true") {
    initialFilterState = {
      ...initialFilterState,
      ...filterConfig(automations, "clusterName"),
    };
  }

  let fields: Field[] = [
    {
      label: "Name",
      value: (k) => {
        const route =
          k.kind === FluxObjectKind.KindKustomization
            ? V2Routes.Kustomization
            : V2Routes.HelmRelease;
        return (
          <Link
            to={formatURL(route, {
              name: k.name,
              namespace: k.namespace,
              clusterName: k.clusterName,
            })}
          >
            {k.name}
          </Link>
        );
      },
      sortValue: ({ name }) => name,
      textSearchable: true,
      maxWidth: 600,
    },
    {
      label: "Type",
      value: "type",
    },
    {
      label: "Namespace",
      value: "namespace",
    },
    ...(flags.WEAVE_GITOPS_FEATURE_TENANCY === "true"
      ? [{ label: "Tenant", value: "tenant" }]
      : []),
    ...(flags.WEAVE_GITOPS_FEATURE_CLUSTER === "true"
      ? [{ label: "Cluster", value: "clusterName" }]
      : []),
    {
      label: "Source",
      value: (a: Automation) => {
        let sourceKind: FluxObjectKind;
        let sourceName: string;
        let sourceNamespace: string;

        if (a.kind === FluxObjectKind.KindKustomization) {
          const sourceRef = getSourceRefForAutomation(a);
          sourceKind = sourceRef?.kind;
          sourceName = sourceRef?.name;
          sourceNamespace = sourceRef?.namespace;
        } else {
          const hr = a as HelmRelease;
          sourceKind = FluxObjectKind.KindHelmChart;
          sourceName = hr.helmChart.name;
          sourceNamespace = hr.helmChart.namespace;
        }

        return (
          <SourceLink
            short
            sourceRef={{
              kind: sourceKind,
              name: sourceName,
              namespace: sourceNamespace,
            }}
            clusterName={a.clusterName}
          />
        );
      },
      sortValue: (a: Automation) => getSourceRefForAutomation(a)?.name,
    },
    {
      label: "Status",
      value: (a: Automation) =>
        a.conditions.length > 0 ? (
          <KubeStatusIndicator
            short
            conditions={a.conditions}
            suspended={a.suspended}
          />
        ) : null,
      sortValue: statusSortHelper,
      defaultSort: true,
    },
    {
      label: "Message",
      value: (a: Automation) => computeMessage(a.conditions),
      sortValue: ({ conditions }) => computeMessage(conditions),
      maxWidth: 600,
    },
    {
      label: "Revision",
      value: "lastAttemptedRevision",
    },
    {
      label: "Last Updated",
      value: (a: Automation) => (
        <Timestamp
          time={_.get(_.find(a.conditions, { type: "Ready" }), "timestamp")}
        />
      ),
      sortValue: (a: Automation) => {
        return _.get(_.find(a.conditions, { type: "Ready" }), "timestamp");
      },
    },
  ];

  if (hideSource) fields = _.filter(fields, (f) => f.label !== "Source");

  return (
    <DataTable
      fields={fields}
      rows={automations}
      className={className}
      filters={initialFilterState}
      hasCheckboxes
    />
  );
}

export default styled(AutomationsTable).attrs({
  className: AutomationsTable.name,
})`
  td:nth-child(7) {
    white-space: pre-wrap;
    overflow-wrap: break-word;
    word-wrap: break-word;
  }
`;
