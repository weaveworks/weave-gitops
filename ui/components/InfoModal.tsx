import { Box, IconButton } from "@material-ui/core";
import { Close } from "@material-ui/icons";
import React, { Dispatch, SetStateAction, useContext } from "react";
import styled from "styled-components";
import { AppContext } from "../contexts/AppContext";
import Flex from "./Flex";
import Text from "./Text";

export type Props = {
  className?: string;
  data: any;
  onClose: Dispatch<SetStateAction<boolean>>;
};

const HeaderFlex = styled(Flex)`
  margin-bottom: ${(props) => props.theme.spacing.xs};
`;

function InfoModal({ data, onClose, className }: Props) {
  return (
    <div className={className}>
      <HeaderFlex wide between align>
        <Text size="large" bold color="neutral30" titleHeight>
          These are the namespaces that we've searched
        </Text>
        <IconButton onClick={() => onClose(false)}>
          <Close />
        </IconButton>
      </HeaderFlex>
      <Box>{data.map((d) => console.log(d))}</Box>
    </div>
  );
}

export default styled(InfoModal).attrs({ className: InfoModal.name })`
  height: 100%;
  padding: ${(props) =>
    props.theme.spacing.small + " " + props.theme.spacing.medium};
`;
