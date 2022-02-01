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
import Button from "./Button";
import Flex from "./Flex";
import Icon, { IconType } from "./Icon";
import Spacer from "./Spacer";
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
  /** A list of strings representing the sortable columns of the table, passed into lodash's `_.sortBy`. Must be lowercase. */
  sortFields?: string[];
  /** an optional list of string widths for each field/column. */
  widths?: string[];
  /** for passing pagination */
  children?: any;
}

const EmptyRow = styled(TableRow)<{ colSpan: number }>`
  td {
    text-align: center;
  }
`;

const TableButton = styled(Button)`
  &.MuiButton-root {
    padding: 0;
    margin: 0;
    text-transform: none;
  }
  &.MuiButton-text {
    min-width: 0px;
    .selected {
      color: ${(props) => props.theme.colors.neutral40};
    }
  }
  &.arrow {
    min-width: 0px;
  }
`;

/** Form DataTable */
function UnstyledDataTable({
  className,
  fields,
  rows,
  sortFields = [],
  widths,
  children,
}: Props) {
  const [sort, setSort] = React.useState(sortFields[0]);
  const [reverseSort, setReverseSort] = React.useState(false);
  const sorted = _.sortBy(rows, sort);

  if (reverseSort) {
    sorted.reverse();
  }

  type labelProps = { label: string };
  function SortableLabel({ label }: labelProps) {
    return (
      <Flex align start>
        <TableButton
          color="inherit"
          variant="text"
          onClick={() => {
            setReverseSort(sort === label.toLowerCase() ? !reverseSort : false);
            setSort(label.toLowerCase());
          }}
        >
          <h2 className={sort === label.toLowerCase() && "selected"}>
            {label}
          </h2>
        </TableButton>
        <Spacer padding="xxs" />
        {sort === label.toLowerCase() ? (
          <Icon
            type={IconType.ArrowUpwardIcon}
            size="base"
            className={reverseSort ? "upward" : "downward"}
          />
        ) : (
          <div style={{ width: "16px" }} />
        )}
      </Flex>
    );
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
                  {sortFields.includes(f.label.toLowerCase()) ? (
                    <SortableLabel label={f.label} />
                  ) : (
                    <h2>{f.label}</h2>
                  )}
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
                  <Flex center align>
                    <Icon
                      color="neutral20"
                      type={IconType.RemoveCircleIcon}
                      size="base"
                    />
                    <Spacer padding="xxs" />
                    <Text color="neutral30">No data</Text>
                  </Flex>
                </TableCell>
              </EmptyRow>
            )}
          </TableBody>
        </Table>
      </TableContainer>
      <Spacer padding="xs" />
      {/* optional pagination component */}
      {children}
    </div>
  );
}

export const DataTable = styled(UnstyledDataTable)`
  h2 {
    font-size: 14px;
    font-weight: 600;
    color: ${(props) => props.theme.colors.neutral30};
    margin: 0px;
  }
  .MuiTableRow-root {
    transition: background 0.5s ease-in-out;
  }
  .MuiTableRow-root:not(.MuiTableRow-head):hover {
    background: ${(props) => props.theme.colors.neutral10};
    transition: background 0.5s ease-in-out;
  }
`;

export default DataTable;
