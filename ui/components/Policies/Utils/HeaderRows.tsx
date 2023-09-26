import React from "react";
import Flex from "../../Flex";
import Text from "../../Text";

export interface RowItem {
  rowkey: string;
  value?: any;
  children?: any;
  visible?: boolean;
}
export function RowHeader({ children, rowkey, value }: RowItem) {
  return (
    <Flex alignItems="center" center gap="8" data-testid={rowkey}>
      <Text color="neutral30" semiBold size="medium" minWidth="150">
        {rowkey}:
      </Text>
      {children ? (
        children
      ) : (
        <Text color="neutral40" size="medium">
          {value || "--"}
        </Text>
      )}
    </Flex>
  );
}
interface Props {
  items: RowItem[];
}
const HeaderRows = ({ items }: Props) => {
  return (
    <Flex column gap="8">
      {items
        .filter((h) => h.visible !== false)
        .map((h) => {
          return (
            <Flex
              alignItems="center"
              center
              gap="8"
              data-testid={h.rowkey}
              key={h.rowkey}
            >
              <Text color="neutral30" semiBold size="medium" minWidth="150">
                {h.rowkey}:
              </Text>
              {h.children ? (
                h.children
              ) : (
                <Text color="neutral40" size="medium">
                  {h.value || "-"}
                </Text>
              )}
            </Flex>
          );
        })}
    </Flex>
  );
};

export default HeaderRows;
