import * as React from "react";
import styled from "styled-components";
/*eslint import/no-unresolved: [0]*/
// @ts-ignore
import logoSrc from "url:../images/logo.svg";
// @ts-ignore
import titleSrc from "url:../images/title.svg";
import Flex from "./Flex";
import Spacer from "./Spacer";

type Props = {
  className?: string;
};

function Logo({ className }: Props) {
  return (
    <Spacer padding="medium">
      <Flex className={className}>
        <img src={logoSrc} />
        <Spacer padding="xxs" />
        <img src={titleSrc} />
      </Flex>
    </Spacer>
  );
}

export default styled(Logo)``;
