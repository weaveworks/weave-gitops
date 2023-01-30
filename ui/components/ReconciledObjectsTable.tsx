import * as React from "react";
import styled from "styled-components";
import { AppContext } from "../contexts/AppContext";
import { useGetReconciledObjects } from "../hooks/flux";
import { Kind } from "../lib/api/core/types.pb";
import { Automation } from "../lib/objects";
import { NoNamespace } from "../lib/types";
import { filterByStatusCallback, filterConfig } from "./DataTable";
import FluxObjectsTable from "./FluxObjectsTable";
import RequestStateHandler from "./RequestStateHandler";

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

  return (
    <RequestStateHandler loading={isLoading} error={error}>
      <FluxObjectsTable
        className={className}
        objects={objs}
        onClick={setNodeYaml}
        initialFilterState={initialFilterState}
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
