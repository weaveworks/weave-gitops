import * as React from "react";
import styled from "styled-components";
/*eslint import/no-unresolved: [0]*/
// @ts-ignore
import logoSrc from "url:../images/logo.svg";
// @ts-ignore
import titleSrc from "url:../images/Title.svg";
import Flex from "./Flex";
import Spacer from "./Spacer";

type Props = {
  className?: string;
};

function Logo({ className }: Props) {
  return (
    <Flex className={className} align start>
      <img src={logoSrc} style={{ height: 56 }} />
      <Spacer padding="xxs" />
      <img src={titleSrc} />
    </Flex>
  );
}

export default styled(Logo)`
  margin-left: ${(props) => props.theme.spacing.medium};
  width: 240px;
`;
