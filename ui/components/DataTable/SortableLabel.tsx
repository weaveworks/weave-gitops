import { Button } from "@material-ui/core";
import * as React from "react";
import styled from "styled-components";
import Flex from "../Flex";
import Icon, { IconType } from "../Icon";
import Spacer from "../Spacer";
import { Field } from "./types";

type labelProps = {
  fields: Field[];
  fieldIndex: number;
  sortFieldIndex: number;
  reverseSort: boolean;
  setSortFieldIndex: (index: number) => void;
  setReverseSort: (b: boolean) => void;
};

const TableButton = styled(Button)`
  &.MuiButton-root {
    margin: 0;
    text-transform: none;
    letter-spacing: 0;
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

export default function SortableLabel({
  fields,
  fieldIndex,
  sortFieldIndex,
  reverseSort,
  setSortFieldIndex,
  setReverseSort,
}: labelProps) {
  const field = fields[fieldIndex];
  const sort = fields[sortFieldIndex];

  return (
    <Flex align start>
      <TableButton
        color="inherit"
        variant="text"
        onClick={() => {
          setReverseSort(sortFieldIndex === fieldIndex ? !reverseSort : false);
          setSortFieldIndex(fieldIndex);
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
