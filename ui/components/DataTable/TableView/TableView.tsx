import { Table, TableContainer } from "@mui/material";
import React from "react";
import TableBodyView from "./TableBody";
import TableHeader from "./TableHeader";
import { TableViewProps } from "./types";

const TableView = ({
  rows,
  fields,
  id,
  checkedFields,
  defaultSortedField,
  onSortChange,
  hasCheckboxes,
  onBatchCheck,
}: TableViewProps) => {
  const onCheckChange = (checked: boolean, id: string) => {
    const selectedFields = [...checkedFields];
    if (checked) {
      selectedFields.push(id);
    } else {
      selectedFields.splice(selectedFields.indexOf(id), 1);
    }
    onBatchCheck(selectedFields);
  };
  const onHeaderCheckChange = (checked: boolean) => {
    if (checked) {
      checkedFields = rows.map((r) => r.uid);
    } else {
      checkedFields = [];
    }
    onBatchCheck(checkedFields);
  };
  return (
    <TableContainer id={id}>
      <Table>
        <TableHeader
          fields={fields}
          defaultSortedField={defaultSortedField}
          hasCheckboxes={hasCheckboxes}
          onSortChange={onSortChange}
          checked={checkedFields.length === rows.length}
          onCheckChange={onHeaderCheckChange}
        />
        <TableBodyView
          rows={rows}
          fields={fields}
          hasCheckboxes={hasCheckboxes}
          onCheckChange={onCheckChange}
          checkedFields={checkedFields}
        />
      </Table>
    </TableContainer>
  );
};

export default TableView;
