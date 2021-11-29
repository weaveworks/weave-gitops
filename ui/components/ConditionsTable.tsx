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
        { label: "Type", value: (c) => c.type },
        { label: "Status", value: (c) => c.status },
        { label: "Reason", value: (c) => c.reason },
        { label: "Message", value: (c) => c.message },
      ]}
      rows={conditions}
      sortFields={[""]}
      className={className}
    />
  );
}

export default styled(ConditionsTable)`
  &.MuiTableCell-head {
    font-weight: 800;
  }
`;
