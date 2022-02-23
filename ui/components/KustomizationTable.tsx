import _ from "lodash";
import * as React from "react";
import styled from "styled-components";
import { Kustomization } from "../lib/api/app/flux.pb";
import { V2Routes } from "../lib/types";
import { computeReady, formatURL } from "../lib/utils";
import DataTable, { SortType } from "./DataTable";
import Link from "./Link";

type Props = {
  className?: string;
  kustomizations: Kustomization[];
};

function KustomizationTable({ className, kustomizations }: Props) {
  return (
    <DataTable
      className={className}
      fields={[
        {
          label: "name",
          value: (k: Kustomization) => (
            <Link
              to={formatURL(V2Routes.Kustomization, {
                name: k.name,
                namespace: k.namespace,
              })}
            >
              {k.name}
            </Link>
          ),
          sortType: SortType.string,
          sortValue: (k: Kustomization) => k.name,
        },
        {
          label: "Ready",
          value: (k: Kustomization) => computeReady(k.conditions),
        },
        {
          label: "Source",
          value: (k: Kustomization) => (
            <Link
              to={formatURL(V2Routes.Source, {
                name: k.sourceRef.name,
                namespace: k.namespace,
                kind: k.sourceRef.kind,
              })}
            >
              {k.sourceRef.name}
            </Link>
          ),
        },
        {
          label: "Message",
          value: (k: Kustomization) => {
            const readyCondition = _.find(
              k.conditions,
              (c) => c.type === "Ready"
            );

            if (readyCondition && readyCondition.status === "False") {
              return readyCondition.message;
            }
          },
        },
        {
          label: "Last Applied Revision",
          value: "lastAppliedRevision",
        },
      ]}
      rows={kustomizations}
    />
  );
}

export default styled(KustomizationTable).attrs({
  className: KustomizationTable.name,
})``;
