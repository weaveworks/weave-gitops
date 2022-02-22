import * as React from "react";
import styled from "styled-components";
import { Automation } from "../hooks/automations";
import { formatURL } from "../lib/nav";
import { AutomationType, V2Routes } from "../lib/types";
import DataTable from "./DataTable";
import Link from "./Link";

type Props = {
  className?: string;
  automations: Automation[];
  appName?: string;
};

function AutomationsTable({ className, automations }: Props) {
  return (
    <DataTable
      className={className}
      sortFields={["name"]}
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
        },
        {
          label: "Type",
          value: "type",
        },
        {
          label: "Namespace",
          value: "namespace",
        },
      ]}
      rows={automations}
    />
  );
}

export default styled(AutomationsTable).attrs({
  className: AutomationsTable.name,
})``;
