import React from "react";
import Flex from "../../Flex";
import Text from "../../Text";

export interface Header {
  rowkey: string;
  value?: any;
  children?: any;
  visible?: boolean;
}

interface Props {
  headers: Header[];
}
const HeaderRows = ({ headers }: Props) => {
  return (
    <Flex column gap="8">
      {headers.map((h) => {
        return (
          h.visible !== false && (
            <Flex align center gap="8" data-testid={h.rowkey} key={h.rowkey}>
              <Text color="neutral30" semiBold size="medium">
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
          )
        );
      })}
    </Flex>
  );
};

export default HeaderRows;
