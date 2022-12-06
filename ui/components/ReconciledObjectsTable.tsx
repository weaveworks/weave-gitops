import * as React from "react";
import styled from "styled-components";
import { AppContext } from "../contexts/AppContext";
import { useLinkResolver } from "../contexts/LinkResolverContext";
import { useGetReconciledObjects } from "../hooks/flux";
import { Kind } from "../lib/api/core/types.pb";
import { formatURL, objectTypeToRoute } from "../lib/nav";
import { Automation, FluxObject } from "../lib/objects";
import { NoNamespace } from "../lib/types";
import { makeImageString, statusSortHelper } from "../lib/utils";
import DataTable, { filterByStatusCallback, filterConfig } from "./DataTable";
import ImageLink from "./ImageLink";
import KubeStatusIndicator, {
  computeMessage,
  createSyntheticCondition,
  ReadyStatusValue,
  SpecialObject,
} from "./KubeStatusIndicator";
import Link from "./Link";
import RequestStateHandler from "./RequestStateHandler";
import Text from "./Text";
interface ReconciledVisualizationProps {
  className?: string;
  automation?: Automation;
}

function ReconciledObjectsTable({
  className,
  automation,
}: ReconciledVisualizationProps) {
  const {
    data: objs,
    error,
    isLoading,
  } = useGetReconciledObjects(
    automation.name,
    automation.namespace || NoNamespace,
    Kind[automation.type],
    automation.inventory,
    automation.clusterName
  );

  const initialFilterState = {
    ...filterConfig(objs, "type"),
    ...filterConfig(objs, "namespace"),
    ...filterConfig(objs, "status", filterByStatusCallback),
  };

  const { setNodeYaml } = React.useContext(AppContext);
  const resolver = useLinkResolver();

  return (
    <RequestStateHandler loading={isLoading} error={error}>
      <DataTable
        filters={initialFilterState}
        className={className}
        fields={[
          {
            value: (u: FluxObject) => {
              const kind = Kind[u.type];
              const secret = u.type === "Secret";
              const params = {
                name: u.name,
                namespace: u.namespace,
                clusterName: u.clusterName,
              };
              // Enterprise is "aware" of more types of objects than Core,
              // and we want to be able to link to those within this table.
              // The resolver func provided by the context will decide what URL this routes to.
              const resolved = resolver && resolver(u.type, params);
              return kind || resolved ? (
                <Link
                  to={resolved || formatURL(objectTypeToRoute(kind), params)}
                >
                  {u.name}
                </Link>
              ) : (
                <Text
                  onClick={() => (secret ? null : setNodeYaml(u))}
                  color={secret ? "neutral40" : "primary10"}
                  pointer={!secret}
                >
                  {u.name}
                </Text>
              );
            },
            label: "Name",
            sortValue: (u: FluxObject) => u.name || "",
            textSearchable: true,
            maxWidth: 600,
          },
          {
            label: "Kind",
            value: (u: FluxObject) => u.type,
            sortValue: (u: FluxObject) => u.type,
          },
          {
            label: "Namespace",
            value: "namespace",
            sortValue: ({ namespace }) => namespace,
          },
          {
            label: "Status",
            value: (u: FluxObject) => {
              const status = u.obj.status;

              if (!status || !status.conditions) {
                const cond = createSyntheticCondition(
                  u.type as SpecialObject,
                  status
                );

                if (cond.status === ReadyStatusValue.Unknown) {
                  return null;
                }

                return (
                  <KubeStatusIndicator
                    conditions={[cond]}
                    suspended={u.suspended}
                    short
                  />
                );
              }

              return u.conditions.length > 0 ? (
                <KubeStatusIndicator
                  conditions={u.conditions}
                  suspended={u.suspended}
                  short
                />
              ) : null;
            },
            sortValue: statusSortHelper,
          },
          {
            label: "Message",
            value: (u: FluxObject) => computeMessage(u.conditions),
            maxWidth: 600,
          },
          {
            label: "Images",
            value: (u: FluxObject) => (
              <ImageLink image={makeImageString(u.images)} />
            ),
            sortValue: (u: FluxObject) => makeImageString(u.images),
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
