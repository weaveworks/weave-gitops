import { Drawer } from "@material-ui/core";
import React, { useContext, useState } from "react";
import styled from "styled-components";
import { AppContext } from "../contexts/AppContext";
import useNavigation from "../hooks/navigation";
import { getParentNavRouteValue } from "../lib/nav";
import { V2Routes } from "../lib/types";
import DetailModal from "./DetailModal";
import Flex from "./Flex";
import { IconType } from "./Icon";
import Logo from "./Logo";
import Nav, { NavItem } from "./Nav";

type Props = {
  className?: string;
  logo?: any;
  nav?: any;
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
];

const AppContainer = styled.div`
  width: 100%;
  min-width: 900px;
  max-width: 100vw;
  min-height: 100vh;
  margin: 0 auto;
  padding: 0;
`;

const Main = styled(Flex)`
  box-sizing: border-box;
`;

function Layout({ className, children }: Props) {
  const [collapsed, setCollapsed] = useState<boolean>(false);

  const { appState, setDetailModal } = useContext(AppContext);
  const detail = appState.detailModal;

  const { currentPage } = useNavigation();
  const value = getParentNavRouteValue(currentPage);

  return (
    <AppContainer className={className}>
      <Main wide tall>
        <Flex column tall>
          <Logo collapsed={collapsed} link={V2Routes.Automations} />
          <Nav
            navItems={navItems}
            collapsed={collapsed}
            setCollapsed={setCollapsed}
            currentPage={value}
          />
        </Flex>
        {children}
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
