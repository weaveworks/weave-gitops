import * as React from "react";
import styled from "styled-components";
import { Automation } from "../hooks/automations";
import { formatURL } from "../lib/nav";
import { AutomationType, V2Routes } from "../lib/types";
import DataTable, { SortType } from "./DataTable";
import Flex from "./Flex";
import KubeStatusIndicator, { computeReady } from "./KubeStatusIndicator";
import Link from "./Link";
import Text from "./Text";

type Props = {
  className?: string;
  automations: Automation[];
  appName?: string;
};

const statusWidth = 360;

function AutomationsTable({ className, automations }: Props) {
  return (
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
        { label: "Last Synced At", value: "lastHandledReconciledAt" },
      ]}
      rows={automations}
    />
  );
}

export default styled(AutomationsTable).attrs({
  className: AutomationsTable.name,
})`
  /* Setting this here to get the ellipsis to work */
  /* Because this is a div within a td, overflow doesn't apply to the td */
  ${KubeStatusIndicator} ${Flex} ${Text} {
    max-width: ${statusWidth}px;
    overflow: hidden;
    text-overflow: ellipsis;
  }
`;
