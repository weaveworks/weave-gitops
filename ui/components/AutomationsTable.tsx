import * as React from "react";
import styled from "styled-components";
import { Button } from "..";
import { Automation } from "../hooks/kustomizations";
import { AutomationKind } from "../lib/api/applications/applications.pb";
import { V2Routes } from "../lib/types";
import { formatAppScopedURL, formatURL } from "../lib/utils";
import DataTable from "./DataTable";
import Flex from "./Flex";
import Link from "./Link";

type Props = {
  className?: string;
  automations: Automation[];
  appName?: string;
};

function AutomationsTable({ className, appName, automations }: Props) {
  return (
    <div className={className}>
      <Flex wide between align>
        <h3>Automations</h3>
        <Link
          to={
            appName
              ? formatAppScopedURL(appName, V2Routes.AddAutomation)
              : formatURL(V2Routes.AddAutomation)
          }
        >
          <Button>Add Automation</Button>
        </Link>
      </Flex>
      <DataTable
        sortFields={["name"]}
        fields={[
          {
            label: "Name",
            value: (k) => {
              const route =
                k.type === AutomationKind.Kustomize
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
    </div>
  );
}

export default styled(AutomationsTable).attrs({
  className: AutomationsTable.name,
})``;
