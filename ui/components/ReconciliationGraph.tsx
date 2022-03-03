import CheckCircleIcon from "@material-ui/icons/CheckCircle";
import ErrorIcon from "@material-ui/icons/Error";
import HourglassFullIcon from "@material-ui/icons/HourglassFull";
import _ from "lodash";
import * as React from "react";
import { renderToString } from "react-dom/server";
import styled from "styled-components";
import {
  UnstructuredObject,
} from "../lib/api/applications/applications.pb";
import { UnstructuredObjectWithParent } from "../lib/graph";
import DirectedGraph from "./DirectedGraph";
import Flex from "./Flex";

export interface Props {
  objects: UnstructuredObjectWithParent[];
  parentObject: any;
  parentObjectKind: string;
  className?: string;
}

function getStatusIcon(status: string) {
  switch (status) {
    case "Current":
      return <CheckCircleIcon />;

    case "InProgress":
      return <HourglassFullIcon />;

    case "Failed":
      return <ErrorIcon color="error" />;

    default:
      return "";
  }
}

type NodeHtmlProps = {
  object: UnstructuredObject;
};

const NodeHtml = ({ object }: NodeHtmlProps) => {
  return (
    <div className="node">
      <Flex center wide align className="name">
        {object.name}
      </Flex>
      <Flex center wide align className="kind">
        <div className="kind-text">{object.groupVersionKind.kind}</div>
      </Flex>
      <Flex center wide align>
        <div className={`status ${object.status}`}>
          {getStatusIcon(object.status)}
        </div>
      </Flex>
    </div>
  );
};

function ReconciliationGraph({
  className,
  objects,
  parentObject,
  parentObjectKind,
}: Props) {
  const edges = _.reduce(
    objects,
    (r, v) => {
      if (v.parentUid) {
        r.push({ source: v.parentUid, target: v.uid });
      } else {
        r.push({ source: parentObject.name, target: v.uid });
      }
      return r;
    },
    []
  );

  const nodes = [
    ..._.map(objects, (r) => ({
      id: r.uid,
      data: r,
      label: (u: UnstructuredObject) => renderToString(<NodeHtml object={u} />),
    })),
    {
      id: parentObject.name,
      data: parentObject,
      label: (u: Props["parentObject"]) =>
        renderToString(
          <NodeHtml
            object={{ ...u, groupVersionKind: { kind: parentObjectKind } }}
          />
        ),
    },
  ];

  return (
    <div className={className}>
      <DirectedGraph
        width="100%"
        height={640}
        scale={1}
        nodes={nodes}
        edges={edges}
        labelShape="ellipse"
        labelType="html"
      />
    </div>
  );
}

export default styled(ReconciliationGraph)`
  ${DirectedGraph} {
    background-color: white;
  }
  .node {
    font-size: 16px;
    /* background-color: white; */
    width: 125px;
    height: 125px;
    display: flex;
    flex-direction: column;
    justify-content: space-evenly;
  }
  ellipse {
    fill: white;
    stroke: #13a000;
    stroke-width: 3;
    stroke-dasharray: 266px;
    filter: drop-shadow(rgb(189, 189, 189) 0px 0px 1px);
  }
  .success ellipse {
    stroke: ${(props) => props.theme.colors.success};
  }
  @keyframes rotate {
    to {
      stroke-dashoffset: 0;
    }
  }
  .status .kind {
    color: ${(props) => props.theme.colors.black};
  }
  .kind-text {
    overflow: hidden;
    text-overflow: ellipsis;
    font-size: 14px;
  }
  .Current {
    color: ${(props) => props.theme.colors.success};
  }
  .name {
    color: ${(props) => props.theme.colors.black};
    font-weight: 800;
    text-align: center;
    white-space: pre-wrap;
  }
  .edgePath path {
    stroke: #bdbdbd;
    stroke-width: 1px;
  }
  .MuiSvgIcon-root {
    height: 24px;
    width: 24px;
  }
`;
