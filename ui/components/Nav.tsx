import { Tab, Tabs } from "@material-ui/core";
import _ from "lodash";
import React from "react";
import styled from "styled-components";
import useNavigation from "../hooks/navigation";
import { formatURL, getParentNavValue } from "../lib/nav";
import { V2Routes } from "../lib/types";
import { colors } from "../typedefs/styled";
import Flex from "./Flex";
import Icon, { IconType } from "./Icon";
import Link from "./Link";
import Spacer from "./Spacer";
import Text from "./Text";

export type NavItem = {
  label: string;
  link: { value: V2Routes | string; href?: string; newTab?: boolean };
  styles: { sub?: boolean; icon?: IconType; groupEnd?: boolean };
};

type Props = {
  className?: string;
  navItems: NavItem[];
};

const NavContent = styled.div`
  //container
  width: 100%;
  height: 100%;
  border-radius: 10px;
  background-color: ${(props) => props.theme.colors.neutral00};
  padding-top: ${(props) => props.theme.spacing.medium};
  box-sizing: border-box;
  //tabs
  .MuiTabs-indicator {
    background-color: ${(props) => props.theme.colors.primary10};
  }
  .link-flex {
    height: 32px;
    padding: 0px 20px;
    margin-bottom: 9px;
    //matches .MuiSvgIcon-root
    transition: background-color 200ms cubic-bezier(0.4, 0, 0.2, 1) 0ms;
    &.group-end {
      margin-bottom: 32px;
    }
    &.selected {
      background-color: rgba(0, 179, 236, 0.1);
    }
    ${Text} {
      letter-spacing: 1px;
    }
  }
`;

const LinkTab = React.forwardRef((p: any, ref) => {
  const item = p.navItem;

  let className = "link-flex";
  if (item.styles.groupEnd) className += " group-end";
  if (p["aria-selected"]) className += " selected";

  let color: keyof typeof colors = "neutral40";
  if (item.styles.sub) color = "neutral30";
  if (p["aria-selected"]) color = "primary10";

  return (
    <Flex wide tall align start className={className} innerRef={ref}>
      {item.styles.icon ? (
        <Icon
          type={item.styles.icon}
          size="medium"
          color={p["aria-selected"] ? "primary10" : "neutral40"}
        />
      ) : (
        <Spacer padding="small" />
      )}
      <Spacer padding="xxs" />
      <Link
        innerRef={ref}
        to={formatURL(item.link.value)}
        href={item.link.href}
        newTab={item.link.newTab}
        textProps={{
          uppercase: true,
          size: "small",
          bold: !item.styles.sub,
          semiBold: item.styles.sub,
          color: color,
        }}
      >
        {...p.children}
      </Link>
    </Flex>
  );
});

function Nav({ className, navItems }: Props) {
  const { currentPage } = useNavigation();
  return (
    <NavContent className={className}>
      <Tabs
        centered={false}
        orientation="vertical"
        value={getParentNavValue(currentPage)}
      >
        {_.map(navItems, (n) => (
          <Tab
            navItem={n}
            key={n.label}
            label={n.label}
            value={n.link.value}
            component={LinkTab}
          />
        ))}
      </Tabs>
    </NavContent>
  );
}

export default styled(Nav).attrs({ className: Nav.name })``;
