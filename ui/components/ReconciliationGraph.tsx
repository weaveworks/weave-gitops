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

  const drag = (simulation) => {
    function dragstarted(event, d) {
      if (!event.active) simulation.alphaTarget(0.3).restart();
      d.fx = d.x;
      d.fy = d.y;
    }

    function dragged(event, d) {
      d.fx = event.x;
      d.fy = event.y;
    }

    function dragended(event, d) {
      if (!event.active) simulation.alphaTarget(0);
      d.fx = null;
      d.fy = null;
    }

    return d3
      .drag()
      .on("start", dragstarted)
      .on("drag", dragged)
      .on("end", dragended);
  };

  type GraphProps = {
    width: number;
    height: number;
    rootNode: UnstructuredObjectWithChildren;
  };

  const MakeGraph = ({ width, height, rootNode }: GraphProps) => {
    const root = d3.hierarchy(rootNode);
    const links = root.links();
    const nodes = root.descendants();
    const svgRef = React.useRef();

    // console.log(root);
    // console.log(nodes);

    const simulation = d3
      .forceSimulation(nodes)
      .force(
        "link",
        d3
          .forceLink(links)
          .id((d) => d.id)
          .distance(0)
          .strength(1)
      )
      .force("charge", d3.forceManyBody().strength(-10000))
      .force("x", d3.forceX())
      .force("y", d3.forceY());

    React.useEffect(() => {
      const svg = d3
        .select(svgRef.current)
        .attr("viewBox", [-width / 2, -height / 2, width, height]);

      const link = svg
        .append("g")
        .attr("stroke", "#999")
        .attr("stroke-opacity", 0.6)
        .selectAll("line")
        .data(links)
        .join("line");

      const node = svg

        .append("g")
        .selectAll("foreignObject")
        .data(nodes)
        .join("foreignObject")

        .html((d) => {
          console.log(d.data);
          const html = renderToString(<NodeHtml object={d.data} />);
          console.log(html);
          return html;
        })

        .call(drag(simulation));

      simulation.on("tick", () => {
        link
          .attr("x1", (d) => d.source.x)
          .attr("y1", (d) => d.source.y)
          .attr("x2", (d) => d.target.x)
          .attr("y2", (d) => d.target.y);

        node.attr("x", (d) => d.x).attr("y", (d) => d.y);
      });
    }, []);

    // invalidation.then(() => simulation.stop());

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
    width: 375px;
    height: 100px;
  }
`;
