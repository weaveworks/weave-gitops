import { Drawer } from "@mui/material";
import React, { useContext, type JSX } from "react";
import styled from "styled-components";
import { AppContext } from "../contexts/AppContext";
import DetailModal from "./DetailModal";
import Flex from "./Flex";

type Props = {
  className?: string;
  logo?: JSX.Element;
  nav?: JSX.Element;
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

function Layout({ className, logo, nav, children }: Props) {
  const { appState, setDetailModal } = useContext(AppContext);
  const detail = appState.detailModal;

  return (
    <AppContainer className={className}>
      <Main wide tall>
        <Flex column tall>
          {logo}
          {nav}
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
