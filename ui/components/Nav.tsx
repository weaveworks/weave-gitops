import { Tab, Tabs } from "@material-ui/core";
import _ from "lodash";
import React, { forwardRef } from "react";
import styled from "styled-components";
import useNavigation from "../hooks/navigation";
import { formatURL, getParentNavValue } from "../lib/nav";
import { V2Routes } from "../lib/types";
import Flex from "./Flex";
import Icon, { IconType } from "./Icon";
import Link from "./Link";
import Spacer from "./Spacer";

export type NavItem = {
  label: string;
  link: { value: V2Routes | string; href?: string; newTab?: boolean };
  styles: { sub?: boolean; icon?: IconType };
};

type Props = {
  className?: string;
  navItems: NavItem[];
};

const NavContent = styled.div`
  height: 100%;
  border-radius: 10px;
  background-color: ${(props) => props.theme.colors.neutral00};
  padding-top: ${(props) => props.theme.spacing.medium};
  padding-left: ${(props) => props.theme.spacing.xs};
  box-sizing: border-box;
  overflow-y: scroll;
  .MuiTabs-indicator {
    background-color: ${(props) => props.theme.colors.primary10};
  }
`;

const LinkTab = styled((navItem: NavItem) => {
  return (
    <Tab
      key={navItem.label}
      component={forwardRef((p: any, ref) => (
        <Flex
          wide
          tall
          align
          start
          key={navItem.label}
          value={navItem.link.value}
          className="link-flex"
        >
          {navItem.styles.icon ? (
            <Icon type={navItem.styles.icon} size="medium" />
          ) : (
            <Spacer padding="small" />
          )}
          <Spacer padding="xxs" />
          <Link
            innerRef={ref}
            to={formatURL(navItem.link.value)}
            href={navItem.link.href}
            newTab={navItem.link.newTab}
            textProps={{
              uppercase: true,
              size: "small",
              bold: !navItem.styles.sub,
              semiBold: navItem.styles.sub,
              color: navItem.styles.sub ? "neutral30" : "neutral40",
            }}
          >
            {navItem.label}
          </Link>
        </Flex>
      ))}
    />
  );
})`
  &.link-flex {
    padding: 9px 20px;
    margin-bottom: ;
  }
`;

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
          <LinkTab navItem={n} />
        ))}
      </Tabs>
    </NavContent>
  );
}

export default styled(Nav).attrs({ className: Nav.name })``;
