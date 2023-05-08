import { Box, IconButton, List, ListItem } from "@material-ui/core";
import React, { Dispatch, SetStateAction, useContext } from "react";
import styled from "styled-components";
import Flex from "./Flex";
import Modal from "./Modal";
import Text from "./Text";

export type Props = {
  data: any;
  onCloseModal: Dispatch<SetStateAction<boolean>>;
  open: boolean;
};

function InfoModal({ data, onCloseModal, open }: Props) {
  const onClose = () => onCloseModal(false);
  const content = (
    <Box>
      <List>
        {data?.map((item) => (
          <ListItem>
            <Text bold>{Object.keys(item)[0]}</Text>:{" "}
            {(Object.values(item)[0] as string[]).join(", ")}
          </ListItem>
        ))}
      </List>
    </Box>
  );
  return (
    <Modal
      open={open}
      onClose={onClose}
      title="Searched namespaces"
      description="These are the namespaces that we've searched per cluster to retrieve the objects that you are seeing on this page."
      children={content}
    />
  );
}

export default InfoModal;
