import _ from "lodash";
import * as React from "react";
import styled from "styled-components";
import { useGetReconciledObjects } from "../hooks/flux";
import {
  AutomationKind,
  GroupVersionKind,
  UnstructuredObject,
} from "../lib/api/core/types.pb";
import { statusSortHelper } from "../lib/utils";
import DataTable, { SortType } from "./DataTable";
import KubeStatusIndicator, { computeMessage } from "./KubeStatusIndicator";
import RequestStateHandler from "./RequestStateHandler";

export interface ReconciledVisualizationProps {
  className?: string;
  automationName: string;
  namespace: string;
  automationKind: AutomationKind;
  kinds: GroupVersionKind[];
  clusterName: string;
}

function ReconciledObjectsTable({
  className,
  automationName,
  namespace,
  automationKind,
  kinds,
  clusterName,
}: ReconciledVisualizationProps) {
  const {
    data: objs,
    error,
    isLoading,
  } = useGetReconciledObjects(
    automationName,
    namespace,
    automationKind,
    kinds,
    clusterName
  );

  return (
    <RequestStateHandler loading={isLoading} error={error}>
      <DataTable
        className={className}
        fields={[
          {
            value: "name",
            label: "Name",
          },
          {
            label: "Type",
            value: (u: UnstructuredObject) => u.groupVersionKind.kind,
            sortType: SortType.string,
            sortValue: (u: UnstructuredObject) => u.groupVersionKind.kind,
          },
          {
            label: "Namespace",
            value: "namespace",
            sortType: SortType.string,
            sortValue: ({ namespace }) => namespace,
          },
          {
            label: "Status",
            value: (u: UnstructuredObject) =>
              u.conditions.length > 0 ? (
                <KubeStatusIndicator
                  conditions={u.conditions}
                  suspended={u.suspended}
                />
              ) : null,
            sortType: SortType.number,
            sortValue: statusSortHelper,
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
    </RequestStateHandler>
  );
}

export default styled(ReconciledObjectsTable).attrs({
  className: ReconciledObjectsTable.name,
})``;
