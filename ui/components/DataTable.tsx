import {
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableRow,
} from "@material-ui/core";
import Checkbox from "@material-ui/core/Checkbox";
import _ from "lodash";
import * as React from "react";
import styled from "styled-components";
import Button from "./Button";
import Flex from "./Flex";
import Icon, { IconType } from "./Icon";

type Props = {
  className?: string;
  fields: {
    label: string;
    value: string | ((k: any) => string | JSX.Element);
  }[];
  rows: any[];
  checks?: boolean;
  sortFields: string[];
  reverseSort?: boolean;
};

const EmptyRow = styled(TableRow)<{ colSpan: number }>`
  font-style: italic;
  td {
    text-align: center;
  }
`;

const TableButton = styled(Button)`
  &.MuiButton-root {
    padding: 0px 4px 0px 0px;
    text-transform: none;
  }
  p {
    margin: 0px;
  }
  &.MuiButton-text {
    min-width: 0px;
  }
  &.bold {
    font-weight: 600;
  }
`;

function DataTable({ className, checks, fields, rows, sortFields }: Props) {
  const [sort, setSort] = React.useState(sortFields[0]);
  const [reverseSort, setReverseSort] = React.useState(false);

  type labelProps = { label: string };
  function SortableLabel({ label }: labelProps) {
    return (
      <Flex align>
        <TableButton
          onClick={() => {
            setSort(label);
            setReverseSort(false);
          }}
          className={`${sort === label && "bold"}`}
        >
          <p>{label}</p>
        </TableButton>
        {sort === label && (
          <TableButton
            onClick={() =>
              reverseSort ? setReverseSort(false) : setReverseSort(true)
            }
          >
            <Icon
              type={IconType.ArrowDownward}
              size="small"
              className={reverseSort ? "upward" : "downward"}
            />
          </TableButton>
        )}
      </Flex>
    );
  }
  const sorted = _.sortBy(rows, sortFields);

  if (reverseSort) {
    sorted.reverse();
  }

  const r = _.map(sorted, (r, i) => (
    <TableRow key={i}>
      {checks && (
        <TableCell key={-1}>
          <Checkbox />
        </TableCell>
      )}
      {_.map(fields, (f, i) => (
        <TableCell key={i}>
          {typeof f.value === "function" ? f.value(r) : r[f.value]}
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
              {checks && (
                <TableCell key={-1}>
                  <Checkbox />
                </TableCell>
              )}
              {_.map(fields, (f, i) => (
                <TableCell key={i}>
                  {sortFields.includes(f.label) ? (
                    <SortableLabel label={f.label} />
                  ) : (
                    f.label
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

export default styled(DataTable)``;
