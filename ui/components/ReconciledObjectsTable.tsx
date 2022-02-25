import _ from "lodash";
import * as React from "react";
import styled from "styled-components";
import { useGetReconciledObjects } from "../hooks/flux";
import {
  AutomationKind,
  GroupVersionKind,
  UnstructuredObject,
} from "../lib/api/core/types.pb";
import DataTable, { SortType } from "./DataTable";
import KubeStatusIndicator, {
  computeMessage,
  computeReady,
} from "./KubeStatusIndicator";

type Props = {
  className?: string;
  automationName: string;
  namespace: string;
  automationKind: AutomationKind;
  kinds: GroupVersionKind[];
};

function ReconciledObjectsTable({
  className,
  automationName,
  namespace,
  automationKind,
  kinds,
}: Props) {
  const objs = useGetReconciledObjects(
    automationName,
    namespace,
    automationKind,
    kinds
  );

  return (
    <div className={className}>
      <DataTable
        fields={[
          {
            value: "name",
            label: "Name",
          },
          {
            label: "Type",
            value: (u: UnstructuredObject) => `${u.groupVersionKind.kind}`,
          },
          {
            label: "Namespace",
            value: "namespace",
          },
          {
            label: "Status",
            value: (u: UnstructuredObject) =>
              u.conditions.length > 0 ? (
                <KubeStatusIndicator conditions={u.conditions} />
              ) : null,
            sortType: SortType.bool,
            sortValue: ({ conditions }) => computeReady(conditions),
          },
          {
            label: "Message",
            value: (u: UnstructuredObject) => _.first(u.conditions)?.message,
            sortType: SortType.string,
            sortValue: ({ conditions }) => computeMessage(conditions),
          },
        ]}
        rows={objs}
      />
    </div>
  );
}

export default styled(ReconciledObjectsTable).attrs({
  className: ReconciledObjectsTable.name,
})``;
