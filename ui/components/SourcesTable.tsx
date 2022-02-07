import * as React from "react";
import styled from "styled-components";
import { Source, V2Routes } from "../lib/types";
import { computeMessage, computeReady, formatURL } from "../lib/utils";
import DataTable from "./DataTable";
import Link from "./Link";

type Props = {
  className?: string;
  sources: Source[];
  appName?: string;
};

function SourcesTable({ className, sources }: Props) {
  return (
    <div className={className}>
      <DataTable
        sortFields={["name"]}
        rows={sources}
        fields={[
          {
            label: "Name",
            value: (k) => (
              <Link
                to={formatURL(V2Routes.Source, {
                  name: k.name,
                  namespace: k.namespace,
                })}
              >
                {k.name}
              </Link>
            ),
          },
          { label: "Namespace", value: "namespace" },
          { label: "Type", value: "type" },
          { label: "Ready", value: (s: Source) => computeReady(s.conditions) },
          {
            label: "Message",
            value: (s: Source) => computeMessage(s.conditions),
          },
        ]}
      />
    </div>
  );
}

export default styled(SourcesTable).attrs({ className: SourcesTable.name })``;
