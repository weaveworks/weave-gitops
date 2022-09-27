import { Dialog } from "@material-ui/core";
import _ from "lodash";
import * as React from "react";
import styled from "styled-components";
import { useGetReconciledObjects } from "../hooks/flux";
import { Kind } from "../lib/api/core/types.pb";
import { formatURL, objectTypeToRoute } from "../lib/nav";
import { Automation, FluxObject } from "../lib/objects";
import { NoNamespace } from "../lib/types";
import { makeImageString, statusSortHelper } from "../lib/utils";
import DataTable, {
  filterByStatusCallback,
  filterByTypeCallback,
  filterConfig,
} from "./DataTable";
import ImageLink from "./ImageLink";
import KubeStatusIndicator, { computeMessage } from "./KubeStatusIndicator";
import Link from "./Link";
import RequestStateHandler from "./RequestStateHandler";
import Text from "./Text";
import { DialogYamlView } from "./YamlView";
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
    ...filterConfig(objs, "type", filterByTypeCallback),
    ...filterConfig(objs, "namespace"),
    ...filterConfig(objs, "status", filterByStatusCallback),
  };

  const [yamlView, setYamlView] = React.useState(null);
  return (
    <RequestStateHandler loading={isLoading} error={error}>
      <DataTable
        filters={initialFilterState}
        className={className}
        fields={[
          {
            value: (u: FluxObject) => {
              const kind = Kind[u.type];
              return kind ? (
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
                <Text onClick={() => setYamlView(u)} color="primary10" pointer>
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
            label: "Type",
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
            value: (u: FluxObject) =>
              u.conditions.length > 0 ? (
                <KubeStatusIndicator
                  conditions={u.conditions}
                  suspended={u.suspended}
                  short
                />
              ) : null,
            sortValue: statusSortHelper,
          },
          {
            label: "Message",
            value: (u: FluxObject) => _.first(u.conditions)?.message,
            sortValue: ({ conditions }) => computeMessage(conditions),
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
      <Dialog open={yamlView !== null} onClose={() => setYamlView(null)}>
        {yamlView && (
          <DialogYamlView
            object={{
              name: yamlView.name,
              namespace: yamlView.namespace,
              kind: yamlView.type,
            }}
            yaml={yamlView.yaml}
          />
        )}
      </Dialog>
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
