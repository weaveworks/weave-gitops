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
import Link from "./Link";
import Spacer from "./Spacer";
import Text from "./Text";

export enum SortType {
  //sort is unused but having number as index zero makes it a falsy value thus not used as a valid sortType for selecting fields for SortableLabel
  sort,
  number,
  string,
  date,
  bool,
}

/** DataTable Properties  */
export interface Props {
  /** CSS MUI Overrides or other styling. */
  className?: string;
  /** A list of objects with four fields: `label`, which is a string representing the column header, `value`, which can be a string, or a function that extracts the data needed to fill the table cell, sortType, which determines the sorting function to be used, and altSortValue, which customizes your input to the search function */
  fields: {
    label: string;
    value: string | ((k: any) => string | JSX.Element);
    sortType?: SortType;
    altSortValue?: (k: any) => any;
  }[];
  /** A list of data that will be iterated through to create the columns described in `fields`. */
  rows: any[];
  /** field to initially sort against. */
  defaultSort: {
    label: string;
    value: string | ((k: any) => string | JSX.Element);
    sortType?: SortType;
    altSortValue?: (k: any) => any;
  };
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

export const sortWithType = (rows, sort) => {
  let altSort;
  if (sort.altSortValue !== undefined) altSort = true;
  switch (sort.sortType) {
    case SortType.number:
      return rows.sort((a: any, b: any) => {
        if (altSort) return sort.altSortValue(a) - sort.altSortValue(b);
        else return a[sort.value] - b[sort.value];
      });

    case SortType.date:
      return rows.sort((a: any, b: any) => {
        if (altSort) {
          return (
            Date.parse(sort.altSortValue(a)) - Date.parse(sort.altSortValue(b))
          );
        } else return a[sort.value] - b[sort.value];
      });

    case SortType.bool:
      return rows.sort((a: any, b: any) => {
        const aVal = altSort ? sort.altSortValue(a) : a[sort.value];
        const bVal = altSort ? sort.altSortValue(b) : b[sort.value];
        if (aVal === bVal) return 0;
        else if (aVal === false && bVal === true) return -1;
        else return 1;
      });

    case SortType.string:
      return altSort
        ? rows.sort((a, b) => {
            if (sort.altSortValue(a) === sort.altSortValue(b)) return 0;
            else if (sort.altSortValue(a) < sort.altSortValue(b)) return -1;
            else return 1;
          })
        : _.sortBy(rows, sort.value);

    default:
      return _.sortBy(rows, sort.value);
  }
};
/** Form DataTable */
function UnstyledDataTable({
  className,
  fields,
  rows,
  defaultSort,
  widths,
  children,
}: Props) {
  const [sort, setSort] = React.useState(defaultSort);
  const [reverseSort, setReverseSort] = React.useState(false);

  const sorted = sortWithType(rows, sort);

  if (reverseSort) {
    sorted.reverse();
  }

  type labelProps = {
    field: {
      label: string;
      value: string | ((k: any) => string | JSX.Element);
      sortType?: SortType;
      altSortValue?: (k: any) => any;
    };
  };
  function SortableLabel({ field }: labelProps) {
    return (
      <Flex align start>
        <TableButton
          color="inherit"
          variant="text"
          onClick={() => {
            setReverseSort(sort === field ? !reverseSort : false);
            setSort(field);
          }}
        >
          <h2 className={sort.label === field.label ? "selected" : ""}>
            {field.label}
          </h2>
        </TableButton>
        <Spacer padding="xxs" />
        {sort.label === field.label ? (
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
                  {f.sortType ? (
                    <SortableLabel field={f} />
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
  ${Link} ${Text} {
    font-size: 14px;
  }
`;

export default DataTable;
