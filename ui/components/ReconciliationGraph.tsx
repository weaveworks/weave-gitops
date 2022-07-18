import * as d3 from "d3";
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

  type GraphProps = {
    width: number;
    height: number;
    rootNode: UnstructuredObjectWithChildren;
  };

  const MakeGraph = ({ width, height, rootNode }: GraphProps) => {
    const svgRef = React.useRef();
    //add source to tree
    const rootNodeWithSource = {
      ...source,
      kind: removeKind(source.kind),
      children: [rootNode],
    };
    const root = d3.hierarchy(rootNodeWithSource);
    const dx = 500;
    const dy = width / 6;
    const tree = d3.tree().nodeSize([dx, dy]);
    const diagonal = d3
      .linkHorizontal()
      .x((d) => d.y)
      .y((d) => d.x);
    const margin = { top: 10, right: 120, bottom: 10, left: 40 };

    root.x0 = dy / 2;
    root.y0 = 0;
    root.descendants().forEach((d, i) => {
      d.id = i;
      d._children = d.children;
      if (d.depth && d.data.name.length !== 7) d.children = null;
    });

    React.useEffect(() => {
      const svg = d3
        .select(svgRef.current)
        .attr("viewBox", [-margin.top, -margin.bottom, width, dx])
        .style("font", "10px sans-serif")
        .style("user-select", "none");

      const gLink = svg
        .append("g")
        .attr("fill", "none")
        .attr("stroke", "#555")
        .attr("stroke-opacity", 0.4)
        .attr("stroke-width", 1.5);

      const gNode = svg
        .append("g")
        .attr("cursor", "pointer")
        .attr("pointer-events", "all");

      function update(source) {
        const duration = d3.event && d3.event.altKey ? 2500 : 250;
        const nodes = root.descendants().reverse();
        const links = root.links();

        // Compute the new tree layout.
        tree(root);

        let left = root;
        let right = root;
        root.eachBefore((node) => {
          if (node.x < left.x) left = node;
          if (node.x > right.x) right = node;
        });

        const height = right.x - left.x + margin.top + margin.bottom;

        const transition = svg
          .transition()
          .duration(duration)
          .attr("viewBox", [-margin.left, left.x - margin.top, width, height])
          .tween(
            "resize",
            window.ResizeObserver ? null : () => () => svg.dispatch("toggle")
          );

        // Update the nodes…
        const node = gNode.selectAll("g").data(nodes, (d) => d.id);

        // Enter any new nodes at the parent's previous position.
        const nodeEnter = node
          .enter()
          .append("g")
          .attr("transform", (d) => `translate(${source.y0},${source.x0})`)
          .attr("fill-opacity", 0)
          .attr("stroke-opacity", 0)
          .on("click", (event, d) => {
            d.children = d.children ? null : d._children;
            update(d);
          });

        const data = nodeEnter.append("foreignObject").html((d) => {
          const html = renderToString(<NodeHtml object={d.data} />);
          return html;
        });

        // Transition nodes to their new position.
        const nodeUpdate = node
          .merge(nodeEnter)
          .transition(transition)
          .attr("transform", (d) => `translate(${d.y},${d.x})`)
          .attr("fill-opacity", 1)
          .attr("stroke-opacity", 1);

        // Transition exiting nodes to the parent's new position.
        const nodeExit = node
          .exit()
          .transition(transition)
          .remove()
          .attr("transform", (d) => `translate(${source.y},${source.x})`)
          .attr("fill-opacity", 0)
          .attr("stroke-opacity", 0);

        // Update the links…
        const link = gLink.selectAll("path").data(links, (d) => d.target.id);

        // Enter any new links at the parent's previous position.
        const linkEnter = link
          .enter()
          .append("path")
          .attr("d", (d) => {
            const o = { x: source.x0, y: source.y0 };
            return diagonal({ source: o, target: o });
          });

        // Transition links to their new position.
        link.merge(linkEnter).transition(transition).attr("d", diagonal);

        // Transition exiting nodes to the parent's new position.
        link
          .exit()
          .transition(transition)
          .remove()
          .attr("d", (d) => {
            const o = { x: source.x, y: source.y };
            return diagonal({ source: o, target: o });
          });

        // Stash the old positions for transition.
        root.eachBefore((d) => {
          d.x0 = d.x;
          d.y0 = d.y;
        });
      }

      update(root);
    }, []);

    return <svg ref={svgRef} preserveAspectRatio="meet" />;
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
    width: 650px;
    height: 200px;
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
    width: 650px;
    height: 200px;
  }
`;
