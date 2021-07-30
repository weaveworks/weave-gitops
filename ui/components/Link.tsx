import * as React from "react";
import { useContext } from "react";
import { Link as RouterLink } from "react-router-dom";
import styled from "styled-components";
import { AppContext } from "../contexts/AppContext";
import Text from "./Text";

type Props = {
  className?: string;
  to: string;
  innerRef?: any;
  children?: any;
};

function Link({ children, to, ...props }: Props) {
  const { linkResolver } = useContext(AppContext);
  return (
    <RouterLink {...props} to={linkResolver(to)}>
      <Text color="primary">{children}</Text>
    </RouterLink>
  );
}

export default styled(Link)`
  text-decoration: none;
`;
