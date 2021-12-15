import * as React from "react";
import styled from "styled-components";
import Button from "./Button";
import Flex from "./Flex";

type Props = {
  className?: string;
  children: any;
};

function ActionBar({ className, children }: Props) {
  return (
    <Flex wide start className={className}>
      {children}
    </Flex>
  );
}

export default styled(ActionBar)`
  margin: 24px 0px;
  ${Button} {
    margin-right: 12px;
  }
`;
