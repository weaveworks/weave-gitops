import * as React from "react";
import styled from "styled-components";
import { AppContext } from "../contexts/AppContext";
import { useGetInventory } from "../hooks/imageautomation";
import { FluxObject } from "../lib/objects";
import { filterByStatusCallback, filterConfig } from "./DataTable";
import FluxObjectsTable from "./FluxObjectsTable";
import RequestStateHandler from "./RequestStateHandler";

interface Props {
  className?: string;
  kind?: string;
  name?: string;
  namespace?: string;
  clusterName?: string;
  withChildren?: boolean;
}

function ReconciledObjectsTable({
  className,
  kind,
  name,
  namespace,
  clusterName,
}: Props) {
  const { data, isLoading, error } = useGetInventory(
    kind,
    name,
    clusterName,
    namespace,
    false,
    {}
  );

  const initialFilterState = {
    ...filterConfig(data?.objects, "type"),
    ...filterConfig(data?.objects, "namespace"),
    ...filterConfig(data?.objects, "status", filterByStatusCallback),
  };
  const { setNodeYaml } = React.useContext(AppContext);
  return (
    <RequestStateHandler loading={isLoading} error={error}>
      <FluxObjectsTable
        className={className}
        objects={data?.objects as FluxObject[]}
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
