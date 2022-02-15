import { Tab, Tabs } from "@material-ui/core";
import _ from "lodash";
import React, { forwardRef } from "react";
import styled from "styled-components";
import useNavigation from "../hooks/navigation";
import images from "../lib/images";
import { V2Routes } from "../lib/types";
import { formatURL, getNavValue } from "../lib/utils";
import { FeatureFlags } from "../contexts/FeatureFlags";
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
    value: V2Routes.ApplicationList,
    label: "Applications",
  },
  {
    value: V2Routes.SourcesList,
    label: "Sources",
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
    color: #4b4b4b;
  }
`;

const negativeSpaceColor = "#f5f5f5";

const AppContainer = styled.div`
  width: 100%;
  overflow-x: hidden;
  height: 100%;
  margin: 0 auto;
  padding: 0;
  background-color: ${negativeSpaceColor};
  background-image: url(${images.background});
  background-position: bottom right;
  background-repeat: no-repeat;
`;

const NavContainer = styled.div`
  width: 240px;
  min-height: 100%;
  margin-top: ${(props) => props.theme.spacing.medium};
  padding-right: ${(props) => props.theme.spacing.small};
  background-color: ${(props) => props.theme.colors.white};
`;

const NavContent = styled.div`
  min-height: 100%;
  padding-top: ${(props) => props.theme.spacing.medium};
  padding-left: ${(props) => props.theme.spacing.xs};
  padding-right: ${(props) => props.theme.spacing.xl};

  .MuiTab-wrapper {
    text-transform: capitalize;
    font-size: 20px;
    font-weight: bold;
  }

  .MuiTabs-indicator {
    left: 0;
    width: 4px;
    background-color: ${(props) => props.theme.colors.primary};
  }

  ${Link} {
    justify-content: flex-start;
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
                value={getNavValue(currentPage)}
              >
                {_.map(navItems, (n) => (
                  <StyleLinkTab
                    key={n.value}
                    label={n.label}
                    to={formatURL(n.value)}
                    value={n.value}
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
