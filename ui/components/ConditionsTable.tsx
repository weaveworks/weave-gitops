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
        { label: "Type", value: "type" },
        { label: "Status", value: "status" },
        { label: "Reason", value: "reason" },
        { label: "Message", value: "message" },
      ]}
      rows={conditions}
      sortFields={[""]}
      className={className}
      widths={["10%", "10%", "25%", "55%"]}
    />
  );
}

export default styled(ConditionsTable)`
  &.MuiTableCell-head {
    font-weight: 800;
  }
`;
