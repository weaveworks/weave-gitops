import * as React from "react";
import styled from "styled-components";
import Flex from "./Flex";
import Text from "./Text";

type Props = {
  className?: string;
  name: string;
  type: string;
};

// eslint-disable-next-line
function DetailTitle({ className, name, type }: Props) {
  //the correct value for the type prop is not currently available - will be added later
  return (
    <Flex align className={className}>
      <Text size="large" semiBold>
        {name}
      </Text>
      {/* <Spacer padding="xs" />
      <Text size="large" color="neutral30">
        {type}
      </Text> */}
    </Flex>
  );
}

export default styled(DetailTitle).attrs({ className: DetailTitle.name })`
  //matches nav
  line-height: 1.75;
`;
