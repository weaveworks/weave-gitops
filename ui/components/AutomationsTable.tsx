import _ from "lodash";
import * as React from "react";
import styled from "styled-components";
import { Automation } from "../hooks/automations";
import { HelmRelease, SourceRefSourceKind } from "../lib/api/core/types.pb";
import { formatURL } from "../lib/nav";
import { AutomationType, V2Routes } from "../lib/types";
import DataTable, { Field, SortType } from "./DataTable";
import FilterableTable, { filterConfigForType } from "./FilterableTable";
import KubeStatusIndicator, {
  computeMessage,
  computeReady,
} from "./KubeStatusIndicator";
import Link from "./Link";
import SourceLink from "./SourceLink";

type Props = {
  className?: string;
  automations: Automation[];
  appName?: string;
  hideSource?: boolean;
};

function AutomationsTable({ className, automations, hideSource }: Props) {
  const initialFilterState = {
    ...filterConfigForType(automations),
  };

  let fields: Field[] = [
    {
      label: "Name",
      value: (k) => {
        const route =
          k.type === AutomationType.Kustomization
            ? V2Routes.Kustomization
            : V2Routes.HelmRelease;
        return (
          <Link
            to={formatURL(route, {
              name: k.name,
              namespace: k.namespace,
            })}
          >
            {k.name}
          </Link>
        );
      },
      sortValue: ({ name }) => name,
      width: 64,
      textSearchable: true,
    },
    {
      label: "Type",
      value: "type",
      width: 96,
    },
    {
      label: "Namespace",
      value: "namespace",
      width: 64,
    },
    {
      label: "Cluster",
      value: () => "Default",
      width: 64,
    },
    {
      label: "Source",
      value: (a: Automation) => {
        let sourceKind;
        let sourceName;

        if (a.type === AutomationType.Kustomization) {
          sourceKind = a.sourceRef.kind;
          sourceName = a.sourceRef.name;
        } else {
          sourceKind = SourceRefSourceKind.HelmChart;
          sourceName = (a as HelmRelease).helmChart.name;
        }

        return (
          <SourceLink sourceRef={{ kind: sourceKind, name: sourceName }} />
        );
      },
      sortValue: (a: Automation) => a.sourceRef?.name,
      width: 160,
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
      sortType: SortType.bool,
      sortValue: ({ conditions }) => computeReady(conditions),
      width: 64,
    },
    {
      label: "Message",
      value: (a: Automation) => computeMessage(a.conditions),
      width: 360,
      sortValue: ({ conditions }) => computeMessage(conditions),
    },
    {
      label: "Revision",
      value: "lastAttemptedRevision",
      width: 72,
    },
    { label: "Last Updated", value: "lastHandledReconciledAt", width: 120 },
  ];

  if (hideSource) fields = _.filter(fields, { label: "Source" });

  return (
    <div className={className}>
      <FilterableTable
        fields={fields}
        filters={initialFilterState}
        rows={automations}
      />
    </div>
  );
}

export default styled(AutomationsTable).attrs({
  className: AutomationsTable.name,
})`
  ${DataTable} table {
    table-layout: fixed;
  }
`;
