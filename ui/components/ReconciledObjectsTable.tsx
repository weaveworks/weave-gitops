import * as React from "react";
import styled from "styled-components";
import { AppContext } from "../contexts/AppContext";
import { useGetInventory } from "../hooks/inventory";
import { ReconciledObjectsAutomation } from "./AutomationDetail";
import { filterByStatusCallback, filterConfig } from "./DataTable";
import FluxObjectsTable from "./FluxObjectsTable";
import RequestStateHandler from "./RequestStateHandler";

interface Props {
  className?: string;
  reconciledObjectsAutomation: ReconciledObjectsAutomation;
}

function ReconciledObjectsTable({
  className,
  reconciledObjectsAutomation,
}: Props) {
  const { type, name, clusterName, namespace } = reconciledObjectsAutomation;
  const { data, isLoading, error } = useGetInventory(
    type,
    name,
    clusterName,
    namespace,
    false
  );

  const initialFilterState = {
    ...filterConfig(data?.objects, "type"),
    ...filterConfig(data?.objects, "namespace"),
    ...filterConfig(data?.objects, "status", filterByStatusCallback),
  };

  const { setDetailModal } = React.useContext(AppContext);

  return (
    <RequestStateHandler loading={isLoading} error={error}>
      <FluxObjectsTable
        className={className}
        objects={data?.objects}
        onClick={setDetailModal}
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
