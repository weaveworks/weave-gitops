import React from "react";
import Flex from "../../Flex";
import Text from "../../Text";

export interface Header {
  rowkey: string;
  value?: any;
  children?: any;
  visible?: boolean;
}
export function RowHeader({ children, rowkey, value }: Header) {
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
  headers: Header[];
}
function HeaderRows({ headers }: Props) {
  return (
    <Flex column gap="8">
      {headers
        .filter((h) => h.visible)
        .map((h) => {
          return (
            <RowHeader rowkey={h.rowkey} value={h.value} key={h.rowkey}>
              {h.children}
            </RowHeader>
          );
        })}
    </Flex>
  );
}

export default HeaderRows;
