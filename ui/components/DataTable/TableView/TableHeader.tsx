import { Checkbox, TableCell, TableHead, TableRow } from "@mui/material";
import React from "react";
import SortableLabelView from "./SortableLabelView";
import { SortField, TableHeaderProps } from "./types";

const TableHeader = ({
  fields,
  hasCheckboxes,
  checked,
  defaultSortedField,
  onSortChange,
  onCheckChange,
}: TableHeaderProps) => {
  const [sortedField, setSortedField] =
    React.useState<SortField>(defaultSortedField);
  return (
    <TableHead>
      <TableRow>
        {hasCheckboxes && (
          <TableCell key={"checkboxes"}>
            <Checkbox
              checked={checked}
              onChange={(e) => {
                onCheckChange(e.target.checked);
              }}
              color="primary"
            />
          </TableCell>
        )}
        {fields.map((f) => (
          <TableCell key={f.label}>
            {typeof f.labelRenderer === "function" ? (
              f.labelRenderer(f)
            ) : (
              <SortableLabelView
                field={f}
                sortedField={sortedField}
                setSortedField={setSortedField}
                onSortClick={(f) => {
                  onSortChange(f);
                }}
              />
            )}
          </TableCell>
        ))}
      </TableRow>
    </TableHead>
  );
};

export default TableHeader;
