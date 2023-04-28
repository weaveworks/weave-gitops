import { Drawer } from "@material-ui/core";
import React, { useContext, useState } from "react";
import styled from "styled-components";
import { AppContext } from "../contexts/AppContext";
import useNavigation from "../hooks/navigation";
import { getParentNavRouteValue } from "../lib/nav";
import { V2Routes } from "../lib/types";
import Breadcrumbs from "./Breadcrumbs";
import DetailModal from "./DetailModal";
import Flex from "./Flex";
import { IconType } from "./Icon";
import Logo from "./Logo";
import Nav, { NavItem } from "./Nav";
import UserSettings from "./UserSettings";

type Props = {
  className?: string;
  children?: any;
};

const navItems: NavItem[] = [
  {
    label: "Applications",
    link: { value: V2Routes.Automations },
    icon: IconType.ApplicationsIcon,
  },
  {
    label: "Sources",
    link: { value: V2Routes.Sources },
    icon: IconType.SourcesIcon,
  },
  {
    label: "Image Automation",
    link: { value: V2Routes.ImageAutomation },
    icon: IconType.ImageAutomationIcon,
  },
  {
    label: "Flux Runtime",
    link: { value: V2Routes.FluxRuntime },
    icon: IconType.FluxIcon,
  },
  { label: "header test" },
  {
    label: "Notifications",
    link: { value: V2Routes.Notifications },
    icon: IconType.NotificationsIcon,
  },
  {
    label: "Docs",
    link: {
      value: "docs",
      href: "https://docs.gitops.weave.works/",
      newTab: true,
    },
    icon: IconType.DocsIcon,
  },
  {
    label: "Notifications",
    link: { value: V2Routes.Notifications },
    icon: IconType.ClustersIcon,
  },
  {
    label: "Notifications",
    link: { value: V2Routes.Notifications },
    icon: IconType.DeliveryIcon,
  },
  {
    label: "Notifications",
    link: { value: V2Routes.Notifications },
    icon: IconType.ExploreIcon,
  },
  {
    label: "Notifications",
    link: { value: V2Routes.Notifications },
    icon: IconType.GitOpsRunIcon,
  },
  {
    label: "Notifications",
    link: { value: V2Routes.Notifications },
    icon: IconType.GitOpsSetsIcon,
  },
  {
    label: "Notifications",
    link: { value: V2Routes.Notifications },
    icon: IconType.PipelinesIcon,
  },
  {
    label: "Notifications",
    link: { value: V2Routes.Notifications },
    icon: IconType.PoliciesIcon,
  },
  {
    label: "Notifications",
    link: { value: V2Routes.Notifications },
    icon: IconType.PolicyConfigsIcon,
  },
  {
    label: "Notifications",
    link: { value: V2Routes.Notifications },
    icon: IconType.SecretsIcon,
  },
  {
    label: "Notifications",
    link: { value: V2Routes.Notifications },
    icon: IconType.TemplatesIcon,
  },
  {
    label: "Notifications",
    link: { value: V2Routes.Notifications },
    icon: IconType.TerraformIcon,
  },
  {
    label: "Notifications",
    link: { value: V2Routes.Notifications },
    icon: IconType.WorkspacesIcon,
  },
];

const AppContainer = styled.div`
  width: 100%;
  min-width: 900px;
  max-width: 100vw;
  min-height: 100vh;
  margin: 0 auto;
  padding: 0;
`;

//top tool bar height needs to match main padding top
const topBarHeight = "60px";

const ContentContainer = styled.div`
  width: 100%;
  min-width: 900px;
  max-width: 100%;
  //without a hard value in the height property, min-height in the Page component doesn't work
  height: 1px;
  min-height: 100%;
  padding-bottom: ${(props) => props.theme.spacing.small};
  padding-right: ${(props) => props.theme.spacing.medium};
  padding-left: ${(props) => props.theme.spacing.medium};
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
  const [collapsed, setCollapsed] = useState<boolean>(false);

  const { appState, setDetailModal } = useContext(AppContext);
  const detail = appState.detailModal;

  const { currentPage } = useNavigation();
  const value = getParentNavRouteValue(currentPage);

  return (
    <AppContainer className={className}>
      <TopToolBar start align wide>
        <Logo collapsed={collapsed} link={V2Routes.Automations} />
        <Breadcrumbs />
        <UserSettings />
      </TopToolBar>
      <Main wide tall>
        <Nav
          navItems={navItems}
          collapsed={collapsed}
          setCollapsed={setCollapsed}
          currentPage={value}
        />
        <ContentContainer>{children}</ContentContainer>
      </Main>
      <Drawer
        anchor="right"
        open={detail ? true : false}
        onClose={() => setDetailModal(null)}
        ModalProps={{ keepMounted: false }}
      >
        {detail && (
          <DetailModal className={detail.className} object={detail.object} />
        )}
      </Drawer>
    </AppContainer>
  );
}

export default styled(Layout)``;
