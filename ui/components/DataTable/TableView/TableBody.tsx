import { Checkbox, TableBody, TableCell, TableRow } from "@mui/material";
import React from "react";
import styled from "styled-components";
import Flex from "../../Flex";
import Icon, { IconType } from "../../Icon";
import Text from "../../Text";
import { TableBodyViewProps } from "./types";

const EmptyRow = styled(TableRow)<{ colSpan: number }>`
  td {
    text-align: center;
  }
`;

const TableBodyView = ({
  rows,
  fields,
  hasCheckboxes,
  checkedFields,
  emptyMessagePlaceholder,
  onCheckChange,
}: TableBodyViewProps) => {
  if (rows.length === 0) {
    const numFields = fields.length + (hasCheckboxes ? 1 : 0);
    return (
      <TableBody>
        <EmptyRow colSpan={numFields}>
          <TableCell colSpan={numFields}>
            <Flex center align gap="8">
              <Icon
                color="neutral20"
                type={IconType.RemoveCircleIcon}
                size="base"
              />
              {emptyMessagePlaceholder || (
                <Text color="neutral30">No data</Text>
              )}
            </Flex>
          </TableCell>
        </EmptyRow>
      </TableBody>
    );
  }

  return (
    <TableBody>
      {rows?.map((r, i) => {
        return (
          <TableRow key={r.uid || i}>
            {hasCheckboxes && (
              <TableCell style={{ padding: "0" }}>
                <Checkbox
                  checked={checkedFields.findIndex((c) => c === r.uid) > -1}
                  onChange={(e) => {
                    onCheckChange(e.target.checked, r.uid);
                  }}
                  color="primary"
                />
              </TableCell>
            )}
            {fields?.map((f) => {
              const style: React.CSSProperties = {
                ...(f.minWidth && { minWidth: f.minWidth }),
                ...(f.maxWidth && { maxWidth: f.maxWidth }),
              };

              return (
                <TableCell
                  style={Object.keys(style).length > 0 ? style : undefined}
                  key={f.label}
                >
                  <Text>
                    {(typeof f.value === "function"
                      ? f.value(r)
                      : r[f.value]) || "-"}
                  </Text>
                </TableCell>
              );
            })}
          </TableRow>
        );
      })}
    </TableBody>
  );
};

export default TableBodyView;
