import * as React from "react";
import styled from "styled-components";
import { AppContext } from "../contexts/AppContext";
import { ReconciledObjectsAutomation } from "./AutomationDetail";
import { filterByStatusCallback, filterConfig } from "./DataTable";
import FluxObjectsTable from "./FluxObjectsTable";
import RequestStateHandler from "./RequestStateHandler";

interface Props {
  className: string;
  reconciledObjectsAutomation: ReconciledObjectsAutomation;
}

function ReconciledObjectsTable({
  className,
  reconciledObjectsAutomation,
}: Props) {
  const { objects, isLoading, error } = reconciledObjectsAutomation;

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
