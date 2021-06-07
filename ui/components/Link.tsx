import * as React from "react";
import { Link as RouterLink } from "react-router-dom";
import styled from "styled-components";
import Text from "./Text";

type Props = {
  className?: string;
  to: string;
  innerRef?: any;
  children?: any;
};

function Link({ children, ...props }: Props) {
  return (
    <RouterLink {...props}>
      <Text color="primary">{children}</Text>
    </RouterLink>
  );
}

export default styled(Link)`
  text-decoration: none;
`;
