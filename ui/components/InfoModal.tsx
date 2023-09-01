import { List, ListItem } from "@material-ui/core";
import React, { Dispatch, Fragment, SetStateAction } from "react";
import styled from "styled-components";
import { SearchedNamespaces } from "../lib/types";
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
      {Object.entries(searchedNamespaces || []).map(
        ([kind, clusterNamespacesList]) => (
          <Fragment key={kind}>
            <h2>kind: {kind}</h2>
            {clusterNamespacesList?.map((clusterNamespaces) => (
              <ListItem key={clusterNamespaces.clusterName}>
                <ClusterName bold>{clusterNamespaces.clusterName}</ClusterName>
                {clusterNamespaces.namespaces.join(", ")}
              </ListItem>
            ))}
          </Fragment>
        )
      )}
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
