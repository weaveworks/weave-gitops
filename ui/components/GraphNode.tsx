import { Tooltip } from "@material-ui/core";
import * as React from "react";
import styled from "styled-components";
import { AppContext } from "../contexts/AppContext";
import { Kind } from "../lib/api/core/types.pb";
import images from "../lib/images";
import { formatURL, objectTypeToRoute } from "../lib/nav";
import { FluxObjectNode } from "../lib/objects";
import Flex from "./Flex";
import { computeReady, ReadyType } from "./KubeStatusIndicator";
import Link from "./Link";
import Text from "./Text";

type Props = {
  className?: string;
  object?: FluxObjectNode;
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

function GraphNode({ className, object }: Props) {
  const { setNodeYaml } = React.useContext(AppContext);
  const status = computeReady(object.conditions);
  const secret = object.type === "Secret";
  return (
    <Node wide tall between className={className}>
      <StatusLine suspended={object.suspended} status={status} />
      <NodeText tall column>
        <Flex start wide align>
          {getStatusIcon(computeReady(object.conditions), object.suspended)}
          <div style={{ padding: 4 }} />
          <Tooltip
            title={object.name.length > 23 ? object.name : ""}
            placement="top"
          >
            {Kind[object.type] ? (
              <div>
                <Link
                  to={formatURL(objectTypeToRoute(Kind[object.type]), {
                    name: object.name,
                    namespace: object.namespace,
                    clusterName: object.clusterName,
                  })}
                  textProps={{ size: "huge", semiBold: object.isCurrentNode }}
                >
                  {object.name}
                </Link>
              </div>
            ) : (
              <Text
                size="huge"
                onClick={() => (secret ? null : setNodeYaml(object))}
                color={secret ? "neutral40" : "primary10"}
                pointer={!secret}
                semiBold={object.isCurrentNode}
              >
                {object.name}
              </Text>
            )}
          </Tooltip>
        </Flex>
        <Kinds start wide align>
          {object.type || ""}
        </Kinds>
        <Kinds start wide align>
          <span>{object.namespace}</span>
        </Kinds>
      </NodeText>
    </Node>
  );
}

export default styled(GraphNode).attrs({ className: GraphNode.name })`
  span,
  a {
    width: 100%;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
`;
