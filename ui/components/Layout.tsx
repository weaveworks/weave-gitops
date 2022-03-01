import { Tab, Tabs } from "@material-ui/core";
import _ from "lodash";
import React, { forwardRef } from "react";
import styled from "styled-components";
import { FeatureFlags } from "../contexts/FeatureFlags";
import useNavigation from "../hooks/navigation";
import { formatURL, getParentNavValue } from "../lib/nav";
import { V2Routes } from "../lib/types";
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
];

const LinkTab = (props) => (
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
  overflow-x: hidden;
  height: 100%;
  margin: 0 auto;
  padding: 0;
`;

const NavContainer = styled.div`
  width: 240px;
  min-height: 100%;
  margin-top: ${(props) => props.theme.spacing.medium};
  background-color: ${(props) => props.theme.colors.neutral00};
  border-radius: 10px;
`;

const NavContent = styled.div`
  min-height: 100%;
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
      margin-bottom: 24px;
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
  min-height: 100vh;
  padding-top: ${(props) => props.theme.spacing.medium};
  padding-bottom: ${(props) => props.theme.spacing.medium};
  padding-right: ${(props) => props.theme.spacing.medium};
  padding-left: ${(props) => props.theme.spacing.medium};
`;

const Main = styled(Flex)`
  height: 100%;
  flex: 1 1 auto;
`;

const TopToolBar = styled(Flex)`
  padding: 8px 0;
  background-color: ${(props) => props.theme.colors.primary};
  width: 100%;
  height: 80px;
  flex: 0 1 auto;

  ${Logo} img {
    width: 70px;
    height: 72.85px;
  }
`;

function Layout({ className, children }: Props) {
  const { authFlag } = React.useContext(FeatureFlags);
  const { currentPage } = useNavigation();
  return (
    <div className={className}>
      <AppContainer>
        <TopToolBar between align>
          <Logo />
          {authFlag ? <UserSettings /> : null}
        </TopToolBar>
        <Main wide>
          <NavContainer>
            <NavContent>
              <Tabs
                centered={false}
                orientation="vertical"
                value={getParentNavValue(currentPage)}
              >
                {_.map(navItems, (n, i) => (
                  <StyleLinkTab
                    key={n.value}
                    label={n.label}
                    to={formatURL(n.value)}
                    value={n.value}
                    className={n.sub && "sub-item"}
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
