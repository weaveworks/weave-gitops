import * as React from "react";
import styled from "styled-components";
import { AppContext } from "../contexts/AppContext";
import { FluxObject } from "../lib/objects";
import { RequestError } from "../lib/types";
import { filterByStatusCallback, filterConfig } from "./DataTable";
import FluxObjectsTable from "./FluxObjectsTable";
import RequestStateHandler from "./RequestStateHandler";

interface ReconciledVisualizationProps {
  className?: string;
  objects: FluxObject[] | undefined[];
  error?: RequestError;
  isLoading?: boolean;
}

function ReconciledObjectsTable({
  className,
  objects,
  error,
  isLoading,
}: ReconciledVisualizationProps) {
  const initialFilterState = {
    ...filterConfig(objects, "type"),
    ...filterConfig(objects, "namespace"),
    ...filterConfig(objects, "status", filterByStatusCallback),
  };

  const { setNodeYaml } = React.useContext(AppContext);

  return (
    <RequestStateHandler loading={isLoading} error={error}>
      <FluxObjectsTable
        className={className}
        objects={objects}
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
