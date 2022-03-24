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
      width: 5,
      textSearchable: true,
    },
    {
      label: "Type",
      value: "type",
      width: 5,
    },
    {
      label: "Namespace",
      value: "namespace",
      width: 5,
    },
    {
      label: "Cluster",
      value: () => "Default",
      width: 5,
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
      width: 10,
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
      sortType: SortType.number,
      sortValue: ({ conditions, suspended }) => {
        if (suspended) return 2;
        if (computeReady(conditions)) return 3;
        else return 1;
      },
      width: 7.5,
    },
    {
      label: "Message",
      value: (a: Automation) => computeMessage(a.conditions),
      width: 37.5,
      sortValue: ({ conditions }) => computeMessage(conditions),
    },
    {
      label: "Revision",
      value: "lastAttemptedRevision",
      width: 15,
    },
    { label: "Last Updated", value: "lastHandledReconciledAt", width: 10 },
  ];

  if (hideSource) fields = _.filter(fields, (f) => f.label !== "Source");

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
