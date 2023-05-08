import { Box, List, ListItem } from "@material-ui/core";
import React, { Dispatch, SetStateAction } from "react";
import Modal from "./Modal";
import Text from "./Text";
import { SearchedNamespaces } from "../hooks/automations";

export type Props = {
  searchedNamespaces: SearchedNamespaces;
  onCloseModal: Dispatch<SetStateAction<boolean>>;
  open: boolean;
};

function InfoModal({ searchedNamespaces, onCloseModal, open }: Props) {
  const onClose = () => onCloseModal(false);

  const content = (
    <Box>
      <List>
        {searchedNamespaces?.map((ns) => (
          <ListItem>
            <Text bold>{Object.keys(ns)[0]}</Text>:{" "}
            {(Object.values(ns)[0] as string[]).join(", ")}
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
