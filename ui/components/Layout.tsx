import { Drawer } from "@material-ui/core";
import React, { useContext } from "react";
import styled from "styled-components";
import { AppContext } from "../contexts/AppContext";
import Breadcrumbs from "./Breadcrumbs";
import DetailModal from "./DetailModal";
import Flex from "./Flex";
import UserSettings from "./UserSettings";

type Props = {
  className?: string;
  logo?: any;
  nav?: any;
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

function Layout({ className, logo, nav, children }: Props) {
  const { appState, setDetailModal } = useContext(AppContext);
  const detail = appState.detailModal;

  return (
    <AppContainer className={className}>
      <TopToolBar start align wide>
        {logo}
        <Breadcrumbs />
        <UserSettings />
      </TopToolBar>
      <Main wide tall>
        {nav}
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
