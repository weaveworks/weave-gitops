import * as React from "react";
import styled from "styled-components";
import { UnstructuredObjectWithChildren } from "../lib/graph";
import images from "../lib/images";
import Flex from "./Flex";
import { computeReady, ReadyType } from "./KubeStatusIndicator";

type Props = {
  className?: string;
  object?: UnstructuredObjectWithChildren & { kind: string };
};

const nodeBorderRadius = "50px";
const titleFontSize = "48px";
const kindFontSize = "32px";

const GraphIcon = styled.img`
  height: ${titleFontSize};
  width: ${titleFontSize};
`;

const Node = styled(Flex)`
  background: white;
  border: 5px solid #7a7a7a;
  border-left: none;
  border-radius: ${nodeBorderRadius};
`;

const NodeText = styled(Flex)`
  width: 90%;
  align-items: flex-start;
  justify-content: space-evenly;
`;

const Title = styled(Flex)`
  font-size: ${titleFontSize};
  text-overflow: ellipsis;
`;

const Kinds = styled(Flex)`
  font-size: ${kindFontSize};
  color: ${(props) => props.theme.colors.neutral30};
  text-overflow: ellipsis;
`;

type StatusLineProps = {
  suspended: boolean;
  status: ReadyType;
};

const StatusLine = styled.div<StatusLineProps>`
  width: 5%;
  height: 100%;
  border-radius: ${nodeBorderRadius} 0 0 ${nodeBorderRadius};
  background-color: ${(props) => {
    if (props.suspended) return props.theme.colors.suspended;
    else if (props.status === ReadyType.Ready)
      return props.theme.colors.success;
    else if (props.status === ReadyType.Reconciling)
      return props.theme.colors.primary10;
    else if (props.status === ReadyType.NotReady)
      return props.theme.colors.alert;
    else return "white";
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
function GraphNode({ className, object }: Props) {
  console.log(object);
  const status = computeReady(object.conditions);
  return (
    <Node wide tall between className={className}>
      <StatusLine suspended={object.suspended} status={status} />
      <NodeText tall column>
        <Title start wide align>
          {getStatusIcon(computeReady(object.conditions), object.suspended)}
          <div style={{ padding: 4 }} />
          {object.name}
        </Title>
        <Kinds start wide align>
          {object.kind || object.groupVersionKind.kind || ""}
        </Kinds>
        <Kinds start wide align>
          {object.namespace}
        </Kinds>
      </NodeText>
    </Node>
  );
}

`
  .status {
    display: flex;
    align-items: center;
  }
  .kind-text {
    overflow: hidden;
    text-overflow: ellipsis;
    font-size: 28px;
  }
  .status-line {
    width: 2.5%;
    border-radius: 10px 0px 0px 10px;
  }
  .Current {
    color: ${(props) => props.theme.colors.success};
    &.status-line {
      background-color: ${(props) => props.theme.colors.success};
    }
  }
  .InProgress {
    color: ${(props) => props.theme.colors.suspended};
    &.status-line {
      background-color: ${(props) => props.theme.colors.suspended};
    }
  }
  .Failed {
    color: ${(props) => props.theme.colors.alert};
    &.status-line {
      background-color: ${(props) => props.theme.colors.alert};
    }
  }
  .name {
    color: ${(props) => props.theme.colors.black};
    font-weight: 800;
    font-size: 28px;
    white-space: pre-wrap;
  }
  .kind {
    color: ${(props) => props.theme.colors.neutral30};
  }
  .edgePath path {
    stroke: ${(props) => props.theme.colors.neutral30};
    stroke-width: 1px;
  }
`;

export default styled(GraphNode).attrs({ className: GraphNode.name })``;
