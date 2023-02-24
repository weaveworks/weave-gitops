import { IconButton, Tab, Tabs, Tooltip } from "@material-ui/core";
import { ArrowLeft, ArrowRight } from "@material-ui/icons";
import _ from "lodash";
import React, { Dispatch, SetStateAction } from "react";
import styled from "styled-components";
import useNavigation from "../hooks/navigation";
import { formatURL, getParentNavRouteValue } from "../lib/nav";
import { V2Routes } from "../lib/types";
// eslint-disable-next-line
import { colors } from "../typedefs/styled";

import Flex from "./Flex";
import Icon, { IconType } from "./Icon";
import Link from "./Link";
import Spacer from "./Spacer";
import Text from "./Text";

export type NavItem = {
  label: string;
  link: { value: V2Routes | string; href?: string; newTab?: boolean };
  styles: { sub?: boolean; icon?: IconType };
};

type Props = {
  className?: string;
  navItems: NavItem[];
  collapsed: boolean;
  setCollapsed: Dispatch<SetStateAction<boolean>>;
};

const fullWidth = "200px";
const collapsedWidth = "64px"; //24px icon, 20px padding left and right
const NavContainer = styled.div<{ collapsed: boolean }>`
  width: ${(props) => (props.collapsed ? collapsedWidth : fullWidth)};
  min-width: ${(props) => (props.collapsed ? collapsedWidth : fullWidth)};
  height: 100%;
  transition: all 0.5s;
`;

const NavContent = styled.div`
  //container
  width: 100%;
  height: 100%;
  border-radius: 10px;
  border-bottom-right-radius: 0px;
  background-color: ${(props) => props.theme.colors.neutral00};
  padding-top: ${(props) => props.theme.spacing.medium};
  box-sizing: border-box;
  //tabs
  .MuiTabs-flexContainerVertical {
    //allows both borders of first nav item to be visible on focus
    padding: ${(props) => props.theme.spacing.xxs} 0;
  }
  .MuiTabs-indicator {
    background-color: ${(props) => props.theme.colors.primary10};
  }
  .link-flex {
    display: flex;
    align-items: center;
    height: 32px;
    padding: 0px 20px;
    margin-bottom: 9px;
    //matches .MuiSvgIcon-root
    transition: background-color 200ms cubic-bezier(0.4, 0, 0.2, 1) 0ms;
    &.selected {
      background-color: rgba(0, 179, 236, 0.1);
    }
    :hover:not(.selected) {
      background-color: ${(props) => props.theme.colors.neutral10};
    }
    ${Text} {
      letter-spacing: 1px;
      //matches MuiIcon
      transition: color 200ms cubic-bezier(0.4, 0, 0.2, 1) 0ms;
    }
  }
`;

const CollapseFlex = styled(Flex)`
  .MuiButtonBase-root {
    margin: 0 18px 0 4px;
  }
  ${Text} {
    opacity: ${(props) => (props.collapsed ? 0 : 1)};
    transition: opacity 200ms cubic-bezier(0.4, 0, 0.2, 1) 0ms;
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
  const [hovered, setHovered] = React.useState(false);
  const item = p.navItem;

  let className = "link-flex";
  if (p["aria-selected"]) className += " selected";

  let color: keyof typeof colors = "neutral40";
  if (p.collapsed) color = "neutral00";
  if (p["aria-selected"] || hovered) {
    if (p.collapsed) color = "primaryLight05";
    else color = "primary10";
  }

  let iconType = item.styles.icon;
  if (iconType === IconType.FluxIcon && (p["aria-selected"] || hovered))
    iconType = IconType.FluxIconHover;

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
          iconType={iconType}
          color={p["aria-selected"] || hovered ? "primary10" : "neutral40"}
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

function Nav({ className, navItems, collapsed, setCollapsed }: Props) {
  const { currentPage } = useNavigation();
  const value = getParentNavRouteValue(currentPage);

  return (
    <NavContainer collapsed={collapsed}>
      <NavContent className={collapsed ? className + " collapsed" : className}>
        <Tabs
          centered={false}
          orientation="vertical"
          value={value === V2Routes.UserInfo ? false : value}
        >
          {_.map(navItems, (n) => {
            const { label, link } = n;
            return (
              <Tab
                navItem={n}
                key={label}
                label={label}
                value={link.value}
                collapsed={collapsed}
                component={LinkTab}
              />
            );
          })}
        </Tabs>
        <CollapseFlex wide align end collapsed={collapsed}>
          <Text uppercase semiBold color="neutral30" size="small">
            collapse
          </Text>
          <IconButton size="small" onClick={() => setCollapsed(!collapsed)}>
            {collapsed ? <ArrowRight /> : <ArrowLeft />}
          </IconButton>
        </CollapseFlex>
      </NavContent>
    </NavContainer>
  );
}

export default styled(Nav).attrs({ className: Nav.name })``;
