import { List, ListItem } from "@material-ui/core";
import React, { Dispatch, SetStateAction } from "react";
import styled from "styled-components";
import { SearchedNamespaces } from "../hooks/automations";
import Modal from "./Modal";
import Text from "./Text";

export type Props = {
  searchedNamespaces: SearchedNamespaces;
  onCloseModal: Dispatch<SetStateAction<boolean>>;
  open: boolean;
};

const ClusterName = styled(Text)`
  padding: ${(props) => props.theme.spacing.base};
`;

function InfoModal({ searchedNamespaces, onCloseModal, open }: Props) {
  const onClose = () => onCloseModal(false);

  const content = (
    <List>
      {searchedNamespaces?.map((ns, index) => (
        <ListItem key={index}>
          <ClusterName bold>{Object.keys(ns)[0]}</ClusterName>
          {(Object.values(ns)[0] as string[]).join(", ")}
        </ListItem>
      ))}
    </List>
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
