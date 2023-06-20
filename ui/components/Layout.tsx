import { Drawer } from "@material-ui/core";
import React, { useContext, useState } from "react";
import styled from "styled-components";
import { AppContext } from "../contexts/AppContext";
import useNavigation from "../hooks/navigation";
import { getParentNavRouteValue } from "../lib/nav";
import DetailModal from "./DetailModal";
import Flex from "./Flex";
import Nav, { NavItem } from "./Nav";
import Logo from "./Logo";

type Props = {
  className?: string;
  logoLink?: string;
  navItems?: NavItem[];
  children?: any;
};

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

function Layout({ className, logoLink, navItems, children }: Props) {
  const [collapsed, setCollapsed] = useState<boolean>(false);
  const { appState, setDetailModal } = useContext(AppContext);
  const detail = appState.detailModal;
  const { currentPage } = useNavigation();
  const value = getParentNavRouteValue(currentPage);

  return (
    <AppContainer className={className}>
      <Main wide tall>
        <Flex column tall>
          <Logo collapsed={collapsed} link={logoLink} />
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
