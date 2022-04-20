import * as React from "react";
import styled from "styled-components";
import { Condition } from "../lib/api/core/types.pb";
import DataTable, { SortType } from "./DataTable";

type Props = {
  className?: string;
  conditions: Condition[];
};

function ConditionsTable({ className, conditions }: Props) {
  return (
    <DataTable
      fields={[
        {
          label: "Type",
          value: "type",
          sortType: SortType.string,
          sortValue: ({ type }) => type,
        },
        { label: "Status", value: "status" },
        { label: "Reason", value: "reason" },
        { label: "Message", value: "message" },
      ]}
      rows={conditions}
      className={className}
    />
  );
}

export default styled(ConditionsTable)`
  &.MuiTableCell-head {
    font-weight: 800;
  }
`;
