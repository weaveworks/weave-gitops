import { Tab, Tabs } from "@material-ui/core";
import _ from "lodash";
import React, { forwardRef } from "react";
import styled from "styled-components";
import useNavigation from "../hooks/navigation";
import { formatURL, getParentNavValue } from "../lib/nav";
import { V2Routes } from "../lib/types";
import Breadcrumbs from "./Breadcrumbs";
import Flex from "./Flex";
import Icon, { IconType } from "./Icon";
import Link from "./Link";
import Logo from "./Logo";
import Spacer from "./Spacer";
import UserSettings from "./UserSettings";

type Props = {
  className?: string;
  children?: any;
};

const navItems = [
  {
    value: V2Routes.Automations,
    label: "Applications",
    icon: IconType.ApplicationsIcon,
  },
  {
    value: V2Routes.Sources,
    label: "Sources",
    sub: true,
  },

  {
    value: V2Routes.FluxRuntime,
    label: "Flux Runtime",
    icon: IconType.FluxIcon,
  },
  {
    value: "docs",
    label: "Docs",
    href: "https://docs.gitops.weave.works/",
    newTab: true,
    icon: IconType.DocsIcon,
  },
];

const LinkTab = styled((props: any) => {
  return (
    <Tab
      component={forwardRef((p: any, ref) => (
        <Flex
          wide
          tall
          align
          start
          key={props.label}
          value={props.value}
          className="link-flex"
        >
          {props.icon ? (
            <Icon type={props.icon} size="medium" />
          ) : (
            <Spacer padding="small" />
          )}
          <Spacer padding="xxs" />
          <Link
            innerRef={ref}
            to={formatURL(props.value)}
            href={props.href}
            newTab={props.newTab}
            textProps={{
              uppercase: true,
              size: "small",
              bold: !props.sub,
              semiBold: props.sub,
            }}
          >
            {props.label}
          </Link>
        </Flex>
      ))}
      {...props}
    />
  );
})`
  &.link-flex {
    padding: 9px 20px;
    letter-spacing: 1;
  }
`;

const AppContainer = styled.div`
  width: 100%;
  min-width: 900px;
  max-width: 100vw;
  min-height: 100vh;
  margin: 0 auto;
  padding: 0;
`;

const navWidth = "200px";
//top tool bar height needs to match main padding top
const topBarHeight = "60px";

const NavContainer = styled.div`
  width: ${navWidth};
  min-width: ${navWidth};
  height: calc(100% - 24px);
  margin-top: ${(props) => props.theme.spacing.small};
  //topBarHeight + small margin
  transform: translateY(72);
`;

const NavContent = styled.div`
  height: 100%;
  border-radius: 10px;
  background-color: ${(props) => props.theme.colors.neutral00};
  padding-top: ${(props) => props.theme.spacing.medium};
  padding-left: ${(props) => props.theme.spacing.xs};
  box-sizing: border-box;
  overflow-y: auto;
`;

const ContentContainer = styled.div`
  width: 100%;
  min-width: 900px;
  max-width: 100%;
  //without a hard value in the height property, min-height in the Page component doesn't work
  height: 1px;
  min-height: 100%;
  padding-top: ${(props) => props.theme.spacing.small};
  padding-bottom: ${(props) => props.theme.spacing.small};
  padding-right: ${(props) => props.theme.spacing.medium};
  padding-left: ${(props) => props.theme.spacing.medium};
  overflow: hidden;
  overflow-y: auto;
  box-sizing: border-box;
`;

const Main = styled(Flex)`
  padding-top: ${topBarHeight};
  box-sizing: border-box;
`;

const TopToolBar = styled(Flex)`
  position: fixed;
  background-color: ${(props) => props.theme.colors.backGrey};
  height: ${topBarHeight};
  min-width: 650px;
  width: 100%;
  ${UserSettings} {
    justify-self: flex-end;
    margin-left: auto;
  }
  //puts it over nav text (must be an mui thing)
  z-index: 2;
`;

function Layout({ className, children }: Props) {
  const { currentPage } = useNavigation();

  return (
    <AppContainer className={className}>
      <TopToolBar start align wide>
        <Logo />
        <Breadcrumbs />
        <UserSettings />
      </TopToolBar>
      <Main wide tall>
        <NavContainer>
          <NavContent>
            <Tabs
              centered={false}
              orientation="vertical"
              value={getParentNavValue(currentPage)}
            >
              {_.map(navItems, (n) => (
                <LinkTab
                  key={n.label}
                  label={n.label}
                  to={formatURL(n.value)}
                  value={n.value}
                  sub={n.sub}
                  href={n.href}
                  newTab={n.newTab}
                  icon={n.icon}
                />
              ))}
            </Tabs>
          </NavContent>
        </NavContainer>
        <ContentContainer>{children}</ContentContainer>
      </Main>
    </AppContainer>
  );
}

export default styled(Layout)``;
