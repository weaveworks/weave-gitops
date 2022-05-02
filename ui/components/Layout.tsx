import { Tab, Tabs } from "@material-ui/core";
import _ from "lodash";
import React, { forwardRef } from "react";
import styled from "styled-components";
import { useFeatureFlags } from "../hooks/featureflags";
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
  min-width: 768px;
  max-width: 100vw;
  min-height: 100vh;
  margin: 0 auto;
  padding: 0;
`;

const NavContainer = styled.div`
  min-width: 200px;
  height: 100%;
  margin-top: ${(props) => props.theme.spacing.medium};
  background-color: ${(props) => props.theme.colors.neutral00};
  border-radius: 10px;
`;

const NavContent = styled.div`
  padding-top: ${(props) => props.theme.spacing.medium};
  padding-left: ${(props) => props.theme.spacing.xs};
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
  max-width: 100%;
  height: 100%;
  padding-top: ${(props) => props.theme.spacing.medium};
  padding-bottom: ${(props) => props.theme.spacing.medium};
  padding-right: ${(props) => props.theme.spacing.medium};
  padding-left: ${(props) => props.theme.spacing.medium};
  overflow: hidden;
`;

const Main = styled(Flex)`
  padding-top: 80px;
`;

const TopToolBar = styled(Flex)`
  position: fixed;
  background-color: ${(props) => props.theme.colors.primary};
  height: 80px;
  min-width: 768px;
  ${UserSettings} {
    justify-self: flex-end;
    margin-left: auto;
  }
  //puts it over nav text (must be an mui thing)
  z-index: 2;
`;

function Layout({ className, children }: Props) {
  const flags = useFeatureFlags();
  const { currentPage } = useNavigation();

  return (
    <div className={className}>
      <AppContainer>
        <TopToolBar start align wide>
          <Logo />
          <Breadcrumbs />
          {flags.WEAVE_GITOPS_AUTH_ENABLED ? <UserSettings /> : null}
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
    </div>
  );
}

export default styled(Layout)`
  display: flex;
`;
