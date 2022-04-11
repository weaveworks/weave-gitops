import * as React from "react";
import styled from "styled-components";
import images from "../lib/images";
import Flex from "./Flex";
import Spacer from "./Spacer";

type Props = {
  className?: string;
};

function Logo({ className }: Props) {
  return (
    <Flex className={className} align start>
      <img src={images.logoSrc} style={{ height: 56 }} />
      <Spacer padding="xxs" />
      <img src={images.titleSrc} />
    </Flex>
  );
}

export default styled(Logo)`
  margin-left: ${(props) => props.theme.spacing.medium};
  width: 240px;
`;
