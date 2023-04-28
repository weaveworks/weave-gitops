import { IconButton, Tab, Tabs, Tooltip } from "@material-ui/core";
import { ArrowLeft, ArrowRight } from "@material-ui/icons";
import _ from "lodash";
import React, { Dispatch, SetStateAction } from "react";
import styled from "styled-components";
import { formatURL } from "../lib/nav";
import { PageRoute, V2Routes } from "../lib/types";
import { Fade } from "../lib/utils";
// eslint-disable-next-line
import { colors } from "../typedefs/styled";

import Flex from "./Flex";
import Icon, { IconType } from "./Icon";
import Link from "./Link";
import Spacer from "./Spacer";
import Text from "./Text";

export type NavItem = {
  label: string;
  link?: { value: V2Routes | string; href?: string; newTab?: boolean };
  icon?: IconType;
  disabled?: boolean;
};

type Props = {
  className?: string;
  navItems: NavItem[];
  collapsed: boolean;
  setCollapsed: Dispatch<SetStateAction<boolean>>;
  currentPage: V2Routes | PageRoute | string | boolean;
};

const fullWidth = "200px";
const collapsedWidth = "64px"; //24px icon, 20px padding left and right
const NavContainer = styled.div<{ collapsed: boolean }>`
  width: ${(props) => (props.collapsed ? collapsedWidth : fullWidth)};
  min-width: ${(props) => (props.collapsed ? collapsedWidth : fullWidth)};
  height: 100%;
  transition: all 0.5s;
`;

const NavContent = styled.div<{ collapsed: boolean }>`
  //container
  display: flex;
  flex-direction: column;
  justify-content: space-between;
  width: 100%;
  height: 100%;
  border-radius: 10px;
  border-bottom-right-radius: 0px;
  background-color: ${(props) => props.theme.colors.neutral00};
  //28px bottom padding is medium + xxs - in the theme these are strings with px at the end so you can't add them together. This lines up with the footer.
  padding: ${(props) => props.theme.spacing.medium} 0px 28px 0px;
  box-sizing: border-box;

  //mui tabs
  .MuiTabs-flexContainerVertical {
    //allows both borders of first nav item to be visible on focus
    padding: ${(props) => props.theme.spacing.xxs} 0;
  }
  .MuiTabs-indicator {
    background-color: ${(props) => props.theme.colors.primary10};
  }

  //navItems
  .link-flex,
  .header {
    height: 32px;
    padding: 0px 20px;
    ${Text} {
      letter-spacing: 1px;
      //matches MuiIcon
      transition: color 200ms cubic-bezier(0.4, 0, 0.2, 1) 0ms;
    }
  }
  .link-flex {
    margin-bottom: ${(props) => props.theme.spacing.xxs};
    display: flex;
    align-items: center;
    //matches .MuiSvgIcon-root
    transition: background-color 200ms cubic-bezier(0.4, 0, 0.2, 1) 0ms;
    &.selected {
      background-color: rgba(0, 179, 236, 0.1);
    }
    :hover:not(.selected) {
      background-color: ${(props) => props.theme.colors.neutral10};
    }
  }
  .header {
    opacity: ${(props) => (props.collapsed ? 0 : 1)};
    transition: opacity 200ms cubic-bezier(0.4, 0, 0.2, 1) 0ms;
    //must match height to center text
    line-height: 32px;
  }
`;

const CollapseButton = styled(IconButton)`
  &.MuiIconButton-root {
    margin: 0 18px 0 4px;
  }
`;

const LinkTabIcon = ({ iconType, color, collapsed, title }) => {
  if (iconType)
    return (
      <Tooltip arrow placement="right" title={collapsed ? title : ""}>
        <div>
          <Icon type={iconType} size="medium" color={color} />
        </div>
      </Tooltip>
    );
  else return <Spacer padding="small" />;
};

const LinkTab = React.forwardRef((p: any, ref) => {
  const [hovered, setHovered] = React.useState<boolean>(false);
  const item: NavItem = p.navItem;

  let className = "link-flex";
  if (p["aria-selected"]) className += " selected";

  let color: keyof typeof colors = "neutral30";
  if (p.collapsed) color = "neutral00";
  if (p["aria-selected"] || hovered) {
    if (p.collapsed) color = "primaryLight05";
    else color = "primary10";
  }

  return (
    <Link
      className={className}
      innerRef={ref}
      to={formatURL(item.link.value)}
      href={item.link.href}
      newTab={item.link.newTab}
      textProps={{
        uppercase: true,
        size: "small",
        bold: true,
        color: color,
      }}
      icon={
        <LinkTabIcon
          iconType={item.icon}
          color={p["aria-selected"] || hovered ? "primary10" : "neutral30"}
          collapsed={p.collapsed}
          title={item.label}
        />
      }
      onMouseEnter={() => setHovered(true)}
      onMouseLeave={() => setHovered(false)}
    >
      {...p.children}
    </Link>
  );
});

function Nav({
  className,
  navItems,
  collapsed,
  setCollapsed,
  currentPage,
}: Props) {
  return (
    <NavContainer collapsed={collapsed}>
      <NavContent className={className} collapsed={collapsed}>
        <Tabs
          centered={false}
          orientation="vertical"
          value={currentPage === V2Routes.UserInfo ? false : currentPage}
          variant="scrollable"
          scrollButtons="off"
        >
          {_.map(navItems, (n) => {
            if (n.disabled) return;
            if (!n.icon && !n.link)
              return (
                <Text
                  uppercase
                  color="neutral40"
                  semiBold
                  className="header"
                  key={n.label}
                >
                  {n.label}
                </Text>
              );

            return (
              <Tab
                navItem={n}
                key={n.label}
                label={n.label}
                value={n.link.value}
                collapsed={collapsed}
                component={LinkTab}
              />
            );
          })}
        </Tabs>
        <Flex wide align end>
          <Fade center fade={collapsed}>
            <Text
              uppercase
              semiBold
              color="neutral30"
              size="small"
              onClick={() => setCollapsed(!collapsed)}
              pointer={!collapsed}
            >
              collapse
            </Text>
          </Fade>
          <CollapseButton size="small" onClick={() => setCollapsed(!collapsed)}>
            {collapsed ? <ArrowRight /> : <ArrowLeft />}
          </CollapseButton>
        </Flex>
      </NavContent>
    </NavContainer>
  );
}

export default styled(Nav).attrs({ className: Nav.name })``;
