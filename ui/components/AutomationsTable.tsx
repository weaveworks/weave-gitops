import _ from "lodash";
import * as React from "react";
import styled from "styled-components";
import { Button, Icon, IconType } from "..";
import { Automation } from "../hooks/automations";
import { formatURL } from "../lib/nav";
import { AutomationType, V2Routes } from "../lib/types";
import DataTable, { SortType } from "./DataTable";
import FilterableTable from "./FilterableTable";
import FilterDialog, { FilterConfig } from "./FilterDialog";
import Flex from "./Flex";
import KubeStatusIndicator, { computeReady } from "./KubeStatusIndicator";
import Link from "./Link";

type Props = {
  className?: string;
  automations: Automation[];
  appName?: string;
};

function AutomationsTable({ className, automations }: Props) {
  const [slid, setSlid] = React.useState(false);
  const typeVals = _.reduce(
    automations,
    (r, v) => {
      const t = v.type;

      if (!_.includes(r, t)) {
        r.push(t);
      }

      return r;
    },
    []
  );

  const initialFilterState = {
    type: typeVals,
  };

  const [filterState, setFilterState] =
    React.useState<FilterConfig>(initialFilterState);

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
    <div className={className}>
      <Flex wide end>
        <Button
          variant="text"
          color="inherit"
          onClick={() => {
            setSlid(!slid);
          }}
        >
          <Icon type={IconType.FilterIcon} size="medium" color="neutral30" />
        </Button>
      </Flex>

      <Flex>
        <FilterableTable
          filters={filterState}
          fields={fields}
          rows={automations}
        />
        <FilterDialog
          onClose={() => setSlid(false)}
          onFilterSelect={(val) => setFilterState(val)}
          filterList={initialFilterState}
          open={slid}
        />
      </Flex>
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
