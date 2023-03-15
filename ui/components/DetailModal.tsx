import { IconButton } from "@material-ui/core";
import { Close } from "@material-ui/icons";
import React, { useContext } from "react";
import styled from "styled-components";
import { AppContext } from "../contexts/AppContext";
import { FluxObject, FluxObjectNode } from "../lib/objects";
import PodPage from "../pages/v2/PodPage";
import Flex from "./Flex";
import Text from "./Text";
import { DialogYamlView } from "./YamlView";

export type DetailViewProps = {
  className?: string;
  object: FluxObject | FluxObjectNode;
};

const HeaderFlex = styled(Flex)`
  margin-bottom: ${(props) => props.theme.spacing.xs};
`;

export enum AltKinds {
  Pod = "Pod",
}

const content = (object) => {
  switch (object.type) {
    case AltKinds.Pod:
      return <PodPage object={object} />;
    default:
      return (
        <DialogYamlView
          object={{
            name: object.name,
            namespace: object.namespace,
            clusterName: object.clusterName,
            kind: object.type,
          }}
          yaml={object.yaml}
        />
      );
  }
};

function DetailModal({ object, className }: DetailViewProps) {
  const { setDetailModal } = useContext(AppContext);
  return (
    <div className={className}>
      <HeaderFlex between align>
        <Text size="large" bold color="neutral30" titleHeight>
          {object.name}
        </Text>
        <IconButton onClick={() => setDetailModal(null)}>
          <Close />
        </IconButton>
      </HeaderFlex>
      {content(object)}
    </div>
  );
}

export default styled(DetailModal).attrs({ className: DetailModal.name })`
  padding: ${(props) => props.theme.spacing.small};
`;
