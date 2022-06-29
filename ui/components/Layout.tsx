import { Tab, Tabs } from "@material-ui/core";
import _ from "lodash";
import React, { forwardRef } from "react";
import styled from "styled-components";
import useNavigation from "../hooks/navigation";
import { formatURL, getParentNavValue } from "../lib/nav";
import { V2Routes } from "../lib/types";
import Breadcrumbs from "./Breadcrumbs";
import Flex from "./Flex";
import Link from "./Link";
import Logo from "./Logo";
import UserSettings from "./UserSettings";

type Props = {
  className?: string;
  children?: any;
};

const navItems = [
  {
    value: V2Routes.Automations,
    label: "Applications",
  },
  {
    value: V2Routes.Sources,
    label: "Sources",
    sub: true,
  },

  {
    value: V2Routes.FluxRuntime,
    label: "Flux Runtime",
  },
  {
    value: "docs",
    label: "Docs",
    href: "https://docs.gitops.weave.works/",
    newTab: true,
  },
];

const LinkTab = (props: any) => (
  <Tab
    component={forwardRef((p: any, ref) => (
      <Link innerRef={ref} {...p} />
    ))}
    {...props}
  />
);

const StyleLinkTab = styled(LinkTab)`
  span {
    align-items: flex-start;
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

//nav width needs to match content margin left.
const navWidth = "200px";
//top tool bar height needs to match main padding top
const topBarHeight = "60px";

const NavContainer = styled.div`
  position: fixed;
  width: ${navWidth};
  //topBarHeight + correct margins of 36px
  height: calc(100% - 84px);
  margin-top: ${(props) => props.theme.spacing.small};
  margin-bottom: ${(props) => props.theme.spacing.small};
`;

const NavContent = styled.div`
  height: 100%;
  border-radius: 10px;
  background-color: ${(props) => props.theme.colors.neutral00};
  padding-top: ${(props) => props.theme.spacing.medium};
  padding-left: ${(props) => props.theme.spacing.xs};
  box-sizing: border-box;
  overflow-y: scroll;
  .MuiTab-textColorInherit {
    opacity: 1;
    .MuiTab-wrapper {
      font-weight: 600;
      font-size: 20px;
      color: ${(props) => props.theme.colors.neutral40};
    }
    &.sub-item {
      opacity: 0.7;
      .MuiTab-wrapper {
        font-weight: 400;
      }
    }
  }
  .MuiTabs-indicator {
    width: 4px;
    background-color: ${(props) => props.theme.colors.primary};
  }
  .MuiTab-root {
    padding: 0px 12px;
    min-height: 24px;
    &.sub-item {
      margin-bottom: 32px;
    }
  }
  ${Link} {
    justify-content: flex-start;
    &.sub-item {
      font-weight: 400;
    }
  }
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
  margin-left: ${navWidth};
  overflow: hidden;
  overflow-y: scroll;
  box-sizing: border-box;
`;

const Main = styled(Flex)`
  padding-top: ${topBarHeight};
  box-sizing: border-box;
`;

const TopToolBar = styled(Flex)`
  position: fixed;
  background-color: ${(props) => props.theme.colors.primary};
  height: ${topBarHeight};
  min-width: 900px;
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
                <StyleLinkTab
                  key={n.label}
                  label={n.label}
                  to={formatURL(n.value)}
                  value={n.value}
                  className={n.sub && "sub-item"}
                  href={n.href}
                  newTab={n.newTab}
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
