import * as React from "react";
import styled from "styled-components";
import Flex from "./Flex";

type Props = {
  className?: string;
  children?: React.ReactNode;
};

const Bar = styled(Flex)`
  button {
    margin: 5px;
  }
`;

function ActionBar({ className, children }: Props) {
  return <Bar className={className}>{children}</Bar>;
}

export default styled(ActionBar).attrs({ className: ActionBar.name })``;
