import CheckCircleIcon from "@material-ui/icons/CheckCircle";
import ErrorIcon from "@material-ui/icons/Error";
import HourglassFullIcon from "@material-ui/icons/HourglassFull";
import _ from "lodash";
import * as React from "react";
import { renderToString } from "react-dom/server";
import styled from "styled-components";
import { useGetReconciledObjects } from "../hooks/flux";
import { UnstructuredObject } from "../lib/api/core/types.pb";
import DirectedGraph from "./DirectedGraph";
import Flex from "./Flex";
import { ReconciledVisualizationProps } from "./ReconciledObjectsTable";
import RequestStateHandler from "./RequestStateHandler";

export type Props = ReconciledVisualizationProps & {
  parentObject: { name?: string; namespace?: string };
};

function getStatusIcon(status: string, suspended: boolean) {
  if (suspended) return <HourglassFullIcon />;
  switch (status) {
    case "Current":
      return <CheckCircleIcon />;

    case "InProgress":
      return <HourglassFullIcon />;

    case "Failed":
      return <ErrorIcon />;

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
      <Flex
        className={`status-line ${
          object.suspended ? "InProgress" : object.status
        }`}
      />
      <Flex column>
        <Flex start wide align className="name">
          {object.name}
        </Flex>
        <Flex start wide align className="kind">
          <div className="kind-text">{object.groupVersionKind.kind}</div>
        </Flex>
        <Flex start wide align>
          <div
            className={`status ${
              object.suspended ? "InProgress" : object.status
            }`}
          >
            {getStatusIcon(object.status, object.suspended)}
          </div>
        </Flex>
      </Flex>
    </div>
  );
};

function ReconciliationGraph({
  className,
  parentObject,
  automationKind,
  kinds,
  clusterName,
}: Props) {
  const {
    data: objects,
    error,
    isLoading,
  } = parentObject ? useGetReconciledObjects(
    parentObject?.name,
    parentObject?.namespace,
    automationKind,
    kinds,
    clusterName
  ) : { data: [], error: null, isLoading: false };

  const edges = _.reduce(
    objects,
    (r, v: any) => {
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
            object={{ ...u, groupVersionKind: { kind: automationKind } }}
          />
        ),
    },
  ];

  return (
    <RequestStateHandler loading={isLoading} error={error}>
      <div className={className} style={{ height: "100%", width: "100%" }}>
        <DirectedGraph
          width="100%"
          height="100%"
          scale={1}
          nodes={nodes}
          edges={edges}
          labelShape="rect"
          labelType="html"
        />
      </div>
    </RequestStateHandler>
  );
}

export default styled(ReconciliationGraph)`
  .node {
    font-size: 16px;
    height: 200px;
    display: flex;
    justify-content: space-evenly;
  }
  rect {
    fill: white;
    stroke: ${(props) => props.theme.colors.neutral20};
    stroke-width: 3;
  }
  .status .kind {
    color: ${(props) => props.theme.colors.black};
  }
  .kind-text {
    overflow: hidden;
    text-overflow: ellipsis;
    font-size: 14px;
  }
  .status-line {
    width: 20px;
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
  .Alert {
    color: ${(props) => props.theme.colors.alert};
    &.status-line {
      background-color: ${(props) => props.theme.colors.alert};
    }
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
