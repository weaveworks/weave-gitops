import * as React from "react";
import type { JSX } from "react";
import { Link as RouterLink } from "react-router";
import styled from "styled-components";
import { isAllowedLink, isHTTP } from "../lib/utils";
import Spacer from "./Spacer";
import Text, { TextProps } from "./Text";

type Props = {
  className?: string;
  to?: string;
  children?: any;
  href?: any;
  target?: any;
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

const Link = React.forwardRef<HTMLAnchorElement, Props>(function Link(
  {
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
  },
  ref
) {
  if ((href && !isAllowedLink(href)) || (!href && !to)) {
    return (
      <Text className={className} {...textProps}>
        {children}
      </Text>
    );
  }

  const txt = (
    <Text color="primary" {...textProps}>
      {children}
    </Text>
  );

  if (href) {
    return (
      <a
        ref={ref}
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

  if (isHTTP(to) || !isAllowedLink(to)) {
    to = new URL("", window.origin + window.location.pathname + to).toString();
  }

  return (
    <RouterLink
      ref={ref}
      onClick={onClick}
      className={className}
      to={to}
      onMouseEnter={onMouseEnter}
      onMouseLeave={onMouseLeave}
      relative="path"
      {...props}
    >
      {icon && <SpacedIcon icon={icon} />}
      {txt}
    </RouterLink>
  );
});

Link.displayName = "Link";

export default styled(Link).attrs({ className: Link.displayName })`
  text-decoration: none;
`;
