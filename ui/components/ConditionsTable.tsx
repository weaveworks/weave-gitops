import * as React from "react";
import styled from "styled-components";
import { Condition } from "../lib/api/applications/applications.pb";
import DataTable from "./DataTable";

type Props = {
  className?: string;
  conditions: Condition[];
};

function ConditionsTable({ className, conditions }: Props) {
  return (
    <DataTable
      fields={[
        { label: "type", displayLabel: "Type", value: "type" },
        { label: "status", displayLabel: "Status", value: "status" },
        { label: "reason", displayLabel: "Reason", value: "reason" },
        { label: "message", displayLabel: "Message", value: "message" },
      ]}
      rows={conditions}
      sortFields={[""]}
      className={className}
      widths={["10%", "10%", "25%", "55%"]}
      disablePagination
    />
  );
}

export default styled(ConditionsTable)`
  &.MuiTableCell-head {
    font-weight: 800;
  }
`;
