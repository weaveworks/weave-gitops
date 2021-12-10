import {
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableRow,
} from "@material-ui/core";
import _ from "lodash";
import * as React from "react";
import styled from "styled-components";
import Text from "./Text";

/** DataTable Properties  */
export interface Props {
  /** CSS MUI Overrides or other styling. */
  className?: string;
  /** A list of objects with two fields: `label`, which is a string representing the column header, and `value`, which can be a string, or a function that extracts the data needed to fill the table cell. */
  fields: {
    label: string;
    value: string | ((k: any) => string | JSX.Element);
  }[];
  /** A list of data that will be iterated through to create the columns described in `fields`. */
  rows: any[];
  /** A list of strings representing the sortable columns of the table, passed into lodash's `_.sortBy`. */
  sortFields: string[];
  /** Indicates whether to reverse the sorted array. */
  reverseSort?: boolean;
  /** an optional list of string widths for each field/column. */
  widths?: string[];
}

const EmptyRow = styled(TableRow)<{ colSpan: number }>`
  font-style: italic;
  td {
    text-align: center;
  }
`;

/** Form DataTable */
function UnstyledDataTable({
  className,
  fields,
  rows,
  sortFields,
  reverseSort,
  widths,
}: Props) {
  const sorted = _.sortBy(rows, sortFields);

  if (reverseSort) {
    sorted.reverse();
  }

  const r = _.map(sorted, (r, i) => (
    <TableRow key={i}>
      {_.map(fields, (f, i) => (
        <TableCell style={widths && { width: widths[i] }} key={f.label}>
          <Text>{typeof f.value === "function" ? f.value(r) : r[f.value]}</Text>
        </TableCell>
      ))}
    </TableRow>
  ));

  return (
    <div className={className}>
      <TableContainer>
        <Table aria-label="simple table">
          <TableHead>
            <TableRow>
              {_.map(fields, (f, i) => (
                <TableCell style={widths && { width: widths[i] }} key={f.label}>
                  {f.label}
                </TableCell>
              ))}
            </TableRow>
          </TableHead>
          <TableBody>
            {r.length > 0 ? (
              r
            ) : (
              <EmptyRow colSpan={fields.length}>
                <TableCell colSpan={fields.length}>
                  <span style={{ fontStyle: "italic" }}>No rows</span>
                </TableCell>
              </EmptyRow>
            )}
          </TableBody>
        </Table>
      </TableContainer>
    </div>
  );
}

export const DataTable = styled(UnstyledDataTable)`
  .MuiTableCell-head {
    font-weight: 800;
  }
`;

export default DataTable;
