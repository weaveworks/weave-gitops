import * as React from "react";
import { Link as RouterLink } from "react-router-dom";
import styled from "styled-components";
import { isAllowedLink } from "../lib/utils";
import Spacer from "./Spacer";
import Text, { TextProps } from "./Text";

type Props = {
  className?: string;
  to?: string;
  innerRef?: any;
  children?: any;
  href?: any;
  newTab?: boolean;
  textProps?: TextProps;
  icon?: JSX.Element;
  onClick?: (ev: any) => void;
  onMouseEnter?: React.EventHandler<React.SyntheticEvent>;
  onMouseLeave?: React.EventHandler<React.SyntheticEvent>;
};

function Link({
  children,
  href,
  className,
  to = "",
  newTab,
  onClick,
  textProps,
  icon,
  onMouseEnter,
  onMouseLeave,
  ...props
}: Props) {
  if (href && !isAllowedLink(href)) {
    return <Text {...textProps}>{children}</Text>;
  }

  const txt = (
    <Text color="primary" {...textProps}>
      {children}
    </Text>
  );

  if (href) {
    return (
      <a
        className={className}
        href={href}
        target={newTab ? "_blank" : ""}
        rel="noreferrer"
        onMouseEnter={onMouseEnter}
        onMouseLeave={onMouseLeave}
      >
        {icon}
        {icon && <Spacer padding="xxs" />}
        {txt}
      </a>
    );
  }

  return (
    <RouterLink
      onClick={onClick}
      className={className}
      to={to}
      onMouseEnter={onMouseEnter}
      onMouseLeave={onMouseLeave}
      {...props}
    >
      {icon}
      {icon && <Spacer padding="xxs" />}
      {txt}
    </RouterLink>
  );
}

export default styled(Link)`
  text-decoration: none;
`;
