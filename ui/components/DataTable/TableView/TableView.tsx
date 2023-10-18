import { Table, TableContainer } from "@material-ui/core";
import React from "react";
import { TableViewProps } from "./modal";
import TableBodyView from "./TableBody";
import TableHeader from "./TableHeader";

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
    const chk = [...checkedFields];
    if (checked) {
      chk.push(id);
    } else {
      chk.splice(chk.indexOf(id), 1);
    }
    onBatchCheck(chk);
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
      <Table aria-label="simple table">
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
