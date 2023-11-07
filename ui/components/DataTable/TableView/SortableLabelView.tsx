import React from "react";
import Flex from "../../Flex";
import Icon, { IconType } from "../../Icon";
import { TableButton } from "../SortableLabel";
import { SortableLabelViewProps } from "./types";

const SortableLabelView = ({
  field,
  onSortClick,
  setSortedField,
  sortedField,
}: SortableLabelViewProps) => {
  return (
    <Flex align start gap="4">
      <TableButton
        color="inherit"
        variant="text"
        onClick={() => {
          setSortedField({
            ...field,
            reverseSort:
              sortedField?.label === field.label
                ? !sortedField.reverseSort
                : false,
          });
          onSortClick({
            ...field,
            reverseSort:
              sortedField?.label === field.label
                ? !sortedField.reverseSort
                : false,
          });
        }}
      >
        <h2 className={sortedField?.label === field.label ? "selected" : ""}>
          {field.label}
        </h2>
      </TableButton>
      {sortedField?.label === field.label ? (
        <Icon
          type={IconType.ArrowUpwardIcon}
          size="base"
          className={sortedField.reverseSort ? "upward" : "downward"}
        />
      ) : (
        <div style={{ width: "16px" }} />
      )}
    </Flex>
  );
};

export default SortableLabelView;
