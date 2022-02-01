import * as React from "react";
import { Link as RouterLink } from "react-router-dom";
import styled from "styled-components";
import Text from "./Text";

type Props = {
  className?: string;
  to?: string;
  innerRef?: any;
  children?: any;
  href?: any;
  newTab?: boolean;
};

function Link({
  children,
  href,
  className,
  to = "",
  newTab,

  ...props
}: Props) {
  const txt = <Text color="primary">{children}</Text>;

  if (href) {
    return (
      <a
        className={className}
        href={href}
        target={newTab ? "_blank" : ""}
        rel="noreferrer"
      >
        {txt}
      </a>
    );
  }

  return (
    <RouterLink className={className} to={to} {...props}>
      {txt}
    </RouterLink>
  );
}

export default styled(Link)`
  text-decoration: none;
  &.title-bar-button {
    width: 250px;
  }
`;
