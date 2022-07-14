import * as d3 from "d3";
import _ from "lodash";
import * as React from "react";
import { renderToString } from "react-dom/server";
import styled from "styled-components";
import { useGetReconciledObjects } from "../hooks/flux";
import {
  Condition,
  ObjectRef,
  UnstructuredObject,
} from "../lib/api/core/types.pb";
import { UnstructuredObjectWithChildren } from "../lib/graph";
import images from "../lib/images";
import { removeKind } from "../lib/utils";
import Flex from "./Flex";
import { computeReady } from "./KubeStatusIndicator";
import { ReconciledVisualizationProps } from "./ReconciledObjectsTable";
import RequestStateHandler from "./RequestStateHandler";

export type Props = ReconciledVisualizationProps & {
  parentObject: {
    name?: string;
    namespace?: string;
    conditions?: Condition[];
    suspended?: boolean;
    children?: UnstructuredObjectWithChildren[];
  };
  source: ObjectRef;
};

const GraphIcon = styled.img`
  height: 16px;
  width: 16px;
`;

function getStatusIcon(status: string, suspended: boolean) {
  if (suspended) return <GraphIcon src={images.suspendedSrc} />;
  switch (status) {
    case "Current":
      return <GraphIcon src={images.successSrc} />;

    case "InProgress":
      return <GraphIcon src={images.suspendedSrc} />;

    case "Failed":
      return <GraphIcon src={images.failedSrc} />;

    default:
      return "";
  }
}

type NodeHtmlProps = {
  object: UnstructuredObject & { kind?: string };
};

const NodeHtml = ({ object }: NodeHtmlProps) => {
  return (
    <div className="node">
      <Flex
        className={`status-line ${
          object.suspended ? "InProgress" : object.status
        }`}
      />
      <Flex column className="nodeText">
        <Flex start wide align className="name">
          <div
            className={`status ${
              object.suspended ? "InProgress" : object.status
            }`}
          >
            {getStatusIcon(object.status, object.suspended)}
          </div>
          <div style={{ padding: 4 }} />
          {object.name}
        </Flex>
        <Flex start wide align className="kind">
          <div className="kind-text">
            {object.groupVersionKind
              ? object.groupVersionKind.kind
              : object.kind}
          </div>
        </Flex>
        <Flex start wide align className="kind">
          <div className="kind-text">{object.namespace}</div>
        </Flex>
      </Flex>
    </div>
  );
};

const findParentStatus = (parent) => {
  if (parent.suspended) return "InProgress";
  if (computeReady(parent.conditions)) return "Current";
  return "Failed";
};

function ReconciliationGraph({
  className,
  parentObject,
  automationKind,
  kinds,
  clusterName,
  source,
}: Props) {
  const {
    data: objects,
    error,
    isLoading,
  } = parentObject
    ? useGetReconciledObjects(
        parentObject?.name,
        parentObject?.namespace,
        automationKind,
        kinds,
        clusterName
      )
    : { data: [], error: null, isLoading: false };

  const rootNode = parentObject;
  rootNode.children = objects;

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

  const sourceId = `source/${source?.namespace}/${source?.name}`;

  const nodes = [
    ..._.map(objects, (r) => ({
      id: r.uid,
      data: r,
      label: (u: UnstructuredObject) => renderToString(<NodeHtml object={u} />),
    })),
    {
      id: parentObject.name,
      data: { ...parentObject, status: findParentStatus(parentObject) },
      label: (u: Props["parentObject"]) =>
        renderToString(
          <NodeHtml
            object={{
              ...u,
              groupVersionKind: { kind: removeKind(automationKind) },
            }}
          />
        ),
    },
    // Add a node to show the source on the graph
    {
      id: sourceId,
      data: {
        ...source,
        kind: removeKind(source.kind),
      },
      label: (s: ObjectRef) =>
        renderToString(
          <NodeHtml
            object={{ ...s, groupVersionKind: { kind: removeKind(s.kind) } }}
          />
        ),
    },
  ];
  // Edge connecting the source to the automation
  edges.push({
    source: sourceId,
    target: parentObject.name,
  });

  type GraphProps = {
    width: number;
    height: number;
    rootNode: UnstructuredObjectWithChildren;
  };

  const MakeGraph = ({ width, height, rootNode }: GraphProps) => {
    const rootNodeWithSource = {
      ...source,
      kind: removeKind(source.kind),
      children: [rootNode],
    };
    const root = d3.hierarchy(rootNodeWithSource);
    const links = root.links();
    const nodes = root.descendants();
    const svgRef = React.useRef();
    const dx = 650;
    const dy = 200;

    const tree = d3.tree().nodeSize([dx, dy]);

    let x0 = Infinity;
    let x1 = -x0;
    tree(root).each((d) => {
      console.log(d);
      if (d.x > x1) x1 = d.x;
      if (d.x < x0) x0 = d.x;
    });

    React.useEffect(() => {
      const svg = d3
        .select(svgRef.current)
        .attr("viewBox", [0, 0, 1000, x1 - x0 + dx * 2]);

      const g = svg
        .append("g")
        .attr("font-family", "sans-serif")
        .attr("font-size", 10)
        .attr("transform", `translate(${0},${dx - x0})`);

      const link = g
        .append("g")
        .attr("stroke", "#999")
        .attr("stroke-opacity", 0.6)
        .selectAll("line")
        .data(links)
        .join("line")
        .attr("x1", (edge) => edge.source.x)
        .attr("x2", (edge) => edge.target.x)
        .attr("y1", (edge) => edge.source.y)
        .attr("y2", (edge) => edge.target.y);

      const node1 = g
        .append("g")
        .attr("stroke-linejoin", "round")
        .attr("stroke-width", 3)
        .selectAll("g")
        .data(root.descendants())
        .join("g")
        .attr("transform", (d) => `translate(${d.x},${d.y})`);

      const node = node1.append("foreignObject").html((d) => {
        const html = renderToString(<NodeHtml object={d.data} />);
        return html;
      });
    }, []);

    return <svg ref={svgRef} />;
  };

  return (
    <RequestStateHandler loading={isLoading} error={error}>
      <div className={className} style={{ height: "100%", width: "100%" }}>
        <MakeGraph width={1000} height={1000} rootNode={rootNode} />
        {/* <DirectedGraph
          scale={defaultScale}
          nodes={nodes}
          edges={edges}
          labelShape="rect"
          labelType="html"
        /> */}
      </div>
    </RequestStateHandler>
  );
}

export default styled(ReconciliationGraph)`
  .node {
    font-size: 8px;
    width: 375px;
    height: 100px;
    display: flex;
    justify-content: space-between;
  }

  rect {
    fill: white;
    stroke: ${(props) => props.theme.colors.neutral20};
    stroke-width: 3;
  }

  .status {
    display: flex;
    align-items: center;
  }

  .kind-text {
    overflow: hidden;
    text-overflow: ellipsis;
    font-size: 14px;
  }

  .status-line {
    width: 2.5%;
    border-radius: 10px 0px 0px 10px;
  }

  .nodeText {
    width: 95%;
    align-items: flex-start;
    justify-content: space-evenly;
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
    font-size: 14px;
    white-space: pre-wrap;
  }

  .kind {
    color: ${(props) => props.theme.colors.neutral30};
  }

  .edgePath path {
    stroke: ${(props) => props.theme.colors.neutral30};
    stroke-width: 1px;
  }

  foreignObject {
    display: flex;
    flex-direction: column;
    width: 375px;
    height: 100px;
  }
`;
