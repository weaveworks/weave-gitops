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

const SpacedIcon = ({ icon }: { icon: JSX.Element }) => (
  <>
    {icon}
    <Spacer padding="xxs" />
  </>
);

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
        {icon && <SpacedIcon icon={icon} />}
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
      {icon && <SpacedIcon icon={icon} />}
      {txt}
    </RouterLink>
  );
}

export default styled(Link)`
  display: flex;
  align-items: center;
  text-decoration: none;
  ${Text} {
    //matches MuiIcon
    transition: color 200ms cubic-bezier(0.4, 0, 0.2, 1) 0ms;
  }
`;
