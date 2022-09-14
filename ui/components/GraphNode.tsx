import { Tooltip } from "@material-ui/core";
import * as React from "react";
import styled from "styled-components";
import { UnstructuredObjectWithChildren } from "../lib/graph";
import images from "../lib/images";
import { FluxObjectNode } from "../lib/objects";
import Flex from "./Flex";
import { computeReady, ReadyType } from "./KubeStatusIndicator";

type DirectedGraphNode = UnstructuredObjectWithChildren & { kind: string } & {
  isCurrentNode?: boolean;
};

type Props = {
  className?: string;
  object?: FluxObjectNode | DirectedGraphNode;
};

const nodeBorderRadius = 30;
const titleFontSize = "48px";
const kindFontSize = "36px";

const GraphIcon = styled.img`
  height: ${titleFontSize};
  width: ${titleFontSize};
  min-height: ${titleFontSize};
  min-width: ${titleFontSize};
`;

const Node = styled(Flex)`
  background: white;
  border: 5px solid ${(props) => props.theme.colors.neutral30};
  border-radius: ${nodeBorderRadius}px;
  user-select: none;
`;

const NodeText = styled(Flex)`
  width: 90%;
  align-items: flex-start;
  justify-content: space-evenly;
`;

const Title = styled(Flex)`
  font-size: ${titleFontSize};
`;

const Kinds = styled(Flex)`
  font-size: ${kindFontSize};
  color: ${(props) => props.theme.colors.neutral30};
`;

type StatusLineProps = {
  suspended: boolean;
  status: ReadyType;
};

const StatusLine = styled.div<StatusLineProps>`
  width: 5%;
  height: 100%;
  border-radius: ${nodeBorderRadius - 4.5}px 0 0 ${nodeBorderRadius - 4.5}px;
  background-color: ${(props) => {
    if (props.suspended) return props.theme.colors.suspended;
    else if (props.status === ReadyType.Ready)
      return props.theme.colors.success;
    else if (props.status === ReadyType.Reconciling)
      return props.theme.colors.primary10;
    else if (props.status === ReadyType.NotReady)
      return props.theme.colors.alert;
    else return "transparent";
  }};
`;

function getStatusIcon(status: ReadyType, suspended: boolean) {
  if (suspended) return <GraphIcon src={images.suspendedSrc} />;
  switch (status) {
    case ReadyType.Ready:
      return <GraphIcon src={images.successSrc} />;

    case ReadyType.Reconciling:
      return <GraphIcon src={images.reconcileSrc} />;

    case ReadyType.NotReady:
      return <GraphIcon src={images.failedSrc} />;

    default:
      return "";
  }
}

const StyledObjectName = styled.span<{ isCurrentNode: boolean }>`
  font-weight: ${(p) => p.isCurrentNode && "600"};
`;

function GraphNode({ className, object }: Props) {
  const status = computeReady(object.conditions);

  const directedGraphNode = object as DirectedGraphNode;

  return (
    <Node wide tall between className={className}>
      <StatusLine suspended={object.suspended} status={status} />
      <NodeText tall column>
        <Title start wide align>
          {getStatusIcon(computeReady(object.conditions), object.suspended)}
          <div style={{ padding: 4 }} />
          <Tooltip
            placement="top"
            title={object.name.length > 23 ? object.name : ""}
          >
            <StyledObjectName isCurrentNode={object.isCurrentNode}>
              {object.name}
            </StyledObjectName>
          </Tooltip>
        </Title>
        <Kinds start wide align>
          {(object as FluxObjectNode).displayKind ||
            directedGraphNode.kind ||
            directedGraphNode.groupVersionKind.kind ||
            ""}
        </Kinds>
        <Kinds start wide align>
          <span>{object.namespace}</span>
        </Kinds>
      </NodeText>
    </Node>
  );
}

export default styled(GraphNode).attrs({ className: GraphNode.name })`
  span {
    width: 100%;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
`;
