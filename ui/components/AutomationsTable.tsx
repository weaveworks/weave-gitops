import * as React from "react";
import styled from "styled-components";
import { Automation } from "../hooks/automations";
import { formatURL } from "../lib/nav";
import { AutomationType, V2Routes } from "../lib/types";
import DataTable, { SortType } from "./DataTable";
import FilterableTable, { filterConfigForType } from "./FilterableTable";
import FilterDialogButton from "./FilterDialogButton";
import Flex from "./Flex";
import KubeStatusIndicator, { computeReady } from "./KubeStatusIndicator";
import Link from "./Link";

type Props = {
  className?: string;
  automations: Automation[];
  appName?: string;
};

function AutomationsTable({ className, automations }: Props) {
  const [filterDialogOpen, setFilterDialog] = React.useState(false);

  const initialFilterState = {
    ...filterConfigForType(automations),
  };

  const fields = [
    {
      label: "Name",
      value: (k) => {
        const route =
          k.type === AutomationType.Kustomization
            ? V2Routes.Kustomization
            : V2Routes.HelmRepo;
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
      sortType: SortType.string,
      sortValue: ({ name }) => name,
      width: 64,
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
      value: "cluster",
      width: 64,
    },
    {
      label: "Status",
      value: (a: Automation) =>
        a.conditions.length > 0 ? (
          <KubeStatusIndicator conditions={a.conditions} />
        ) : null,
      sortType: SortType.bool,
      sortValue: ({ conditions }) => computeReady(conditions),
      width: 360,
    },
    {
      label: "Revision",
      value: "lastAttemptedRevision",
      width: 72,
    },
    { label: "Last Synced At", value: "lastHandledReconciledAt", width: 120 },
  ];

  return (
<<<<<<< HEAD
    <DataTable
      className={className}
      fields={[
        {
          label: "Name",
          value: (k) => {
            const route =
              k.type === AutomationType.Kustomization
                ? V2Routes.Kustomization
                : V2Routes.HelmRepo;
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
          sortType: SortType.string,
          sortValue: ({ name }) => name,
        },
        {
          label: "Type",
          value: "type",
        },
        {
          label: "Namespace",
          value: "namespace",
        },
        {
          label: "Cluster",
          value: "cluster",
        },
        {
          label: "Status",
          value: (a: Automation) =>
            a.conditions.length > 0 ? (
              <KubeStatusIndicator conditions={a.conditions} />
            ) : null,
          sortType: SortType.bool,
          sortValue: ({ conditions }) => computeReady(conditions),
          width: statusWidth,
        },
        {
          label: "Revision",
          value: "lastAttemptedRevision",
        },
        { label: "Last Updated", value: "lastHandledReconciledAt" },
      ]}
      rows={automations}
    />
=======
    <div className={className}>
      <Flex wide end>
        <FilterDialogButton
          onClick={() => setFilterDialog(!filterDialogOpen)}
        />
      </Flex>

      <FilterableTable
        fields={fields}
        filters={initialFilterState}
        rows={automations}
        dialogOpen={filterDialogOpen}
        onDialogClose={() => setFilterDialog(false)}
      />
    </div>
>>>>>>> v2
  );
}

export default styled(AutomationsTable).attrs({
  className: AutomationsTable.name,
})`
  ${DataTable} table {
    table-layout: fixed;
  }
`;
