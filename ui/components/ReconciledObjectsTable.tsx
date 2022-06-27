import _ from "lodash";
import * as React from "react";
import styled from "styled-components";
import { useGetReconciledObjects } from "../hooks/flux";
import {
  FluxObjectKind,
  GroupVersionKind,
  UnstructuredObject,
} from "../lib/api/core/types.pb";
import { formatURL, objectTypeToRoute } from "../lib/nav";
import { NoNamespace } from "../lib/types";
import { addKind, makeImageString, statusSortHelper } from "../lib/utils";
import { SortType } from "./DataTable";
import {
  filterByStatusCallback,
  filterByTypeCallback,
  filterConfig,
} from "./FilterableTable";
import ImageLink from "./ImageLink";
import KubeStatusIndicator, { computeMessage } from "./KubeStatusIndicator";
import Link from "./Link";
import RequestStateHandler from "./RequestStateHandler";
import URLAddressableTable from "./URLAddressableTable";

export interface ReconciledVisualizationProps {
  className?: string;
  automationName: string;
  namespace?: string;
  automationKind: FluxObjectKind;
  kinds: GroupVersionKind[];
  clusterName: string;
}

const kindsFrom = [
  FluxObjectKind.KindKustomization,
  FluxObjectKind.KindHelmRelease,
];

const kindsTo = [
  FluxObjectKind.KindKustomization,
  FluxObjectKind.KindHelmRelease,
  FluxObjectKind.KindGitRepository,
  FluxObjectKind.KindHelmRepository,
  FluxObjectKind.KindBucket,
];

function ReconciledObjectsTable({
  className,
  automationName,
  namespace = NoNamespace,
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

  const initialFilterState = {
    ...filterConfig(objs, "type", filterByTypeCallback),
    ...filterConfig(objs, "namespace"),
    ...filterConfig(objs, "status", filterByStatusCallback),
  };

  const shouldDisplayLinks = kindsFrom.includes(automationKind);

  return (
    <RequestStateHandler loading={isLoading} error={error}>
      <URLAddressableTable
        filters={initialFilterState}
        className={className}
        fields={[
          {
            value: (u: UnstructuredObject) => {
              const kind = FluxObjectKind[addKind(u.groupVersionKind.kind)];

              return shouldDisplayLinks && kind && kindsTo.includes(kind) ? (
                <Link
                  to={formatURL(objectTypeToRoute(kind), {
                    name: u.name,
                    namespace: u.namespace,
                    clusterName: u.clusterName,
                  })}
                >
                  {u.name}
                </Link>
              ) : (
                u.name
              );
            },
            label: "Name",
            sortValue: (u: UnstructuredObject) => u.name || "",
            textSearchable: true,
            maxWidth: 600,
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
                  short
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
            maxWidth: 600,
          },
          {
            label: "Images",
            value: (u: UnstructuredObject) => (
              <ImageLink image={makeImageString(u.images)} />
            ),
            sortType: SortType.string,
            sortValue: (u: UnstructuredObject) => makeImageString(u.images),
          },
        ]}
        rows={objs}
      />
    </RequestStateHandler>
  );
}

export default styled(ReconciledObjectsTable).attrs({
  className: ReconciledObjectsTable.name,
})`
  td:nth-child(5),
  td:nth-child(6) {
    white-space: pre-wrap;
  }
  td:nth-child(5) {
    overflow-wrap: break-word;
    word-wrap: break-word;
  }
`;
