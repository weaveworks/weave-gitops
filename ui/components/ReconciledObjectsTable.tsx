import * as React from "react";
import styled from "styled-components";
import { AppContext } from "../contexts/AppContext";
import { FluxObject } from "../lib/objects";
import { filterByStatusCallback, filterConfig } from "./DataTable";
import FluxObjectsTable from "./FluxObjectsTable";

interface Props {
  className?: string;
  objects: FluxObject[];
}

function ReconciledObjectsTable({ className, objects }: Props) {
  const initialFilterState = {
    ...filterConfig(objects, "type"),
    ...filterConfig(objects, "namespace"),
    ...filterConfig(objects, "status", filterByStatusCallback),
  };

  const { setDetailModal } = React.useContext(AppContext);

  return (
    <FluxObjectsTable
      className={className}
      objects={objects}
      onClick={setDetailModal}
      initialFilterState={initialFilterState}
    />
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
