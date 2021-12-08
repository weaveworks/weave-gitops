import { Slider } from "@material-ui/core";
import * as d3 from "d3";
import dagreD3 from "dagre-d3";
import _ from "lodash";
import * as React from "react";
import styled from "styled-components";

type Props<N> = {
  className?: string;
  nodes: { id: any; data: N; label: (v: N) => string }[];
  edges: { source: any; target: any }[];
  scale?: number;
  width: number | string;
  height: number;
  labelType?: "html" | "text";
  labelShape: "rect" | "ellipse";
};

const GraphSlider = styled(Slider)`
  &.MuiSlider-root {
    &.MuiSlider-vertical {
      position: relative;
      height: 150px;
      left: 95%;
      top: 5vh;
    }
  }
`;

function DirectedGraph<T>({
  className,
  nodes,
  edges,
  scale,
  width,
  height,
  labelType,
  labelShape,
}: Props<T>) {
  const svgRef = React.useRef();

  const [zoomPercent, setZoomPercent] = React.useState(0);

  React.useEffect(() => {
    if (!svgRef.current) {
      return;
    }

    // https://github.com/jsdom/jsdom/issues/2531
    if (process.env.NODE_ENV === "test") {
      return;
    }

    const dagreD3LibRef = dagreD3;
    const graph = new dagreD3LibRef.graphlib.Graph();

    graph
      .setGraph({
        nodesep: 50,
        ranksep: 50,
        rankdir: "TB",
      })
      .setDefaultEdgeLabel(() => {
        return {};
      });

    _.each(nodes, (n) => {
      graph.setNode(n.id, {
        label: n.label(n.data),
        labelType,
        shape: labelShape,
        width: 150,
        height: 150,
        labelClass: "foo",
      });
    });

    _.each(edges, (e) => {
      graph.setEdge(e.source, e.target, { arrowhead: "undirected" });
    });

    // Create the renderer
    const render = new dagreD3.render();

    // Set up an SVG group so that we can translate the final graph.
    const svg = d3.select(svgRef.current);
    svg.append("g");

    // Set up zoom support
    const zoom = d3.zoom().on("zoom", (e) => {
      e.transform.k = zoomPercent / 100;
      svg.select("g").attr("transform", e.transform);
    });

    svg
      .call(zoom)
      .call(zoom.transform, d3.zoomIdentity.scale(scale))
      .on("wheel.zoom", null);

    // Run the renderer. This is what draws the final graph.
    render(d3.select("svg g"), graph);
  }, [svgRef.current, nodes, edges, zoomPercent]);
  return (
    <div className={className}>
      <GraphSlider
        onChange={(e, value: number) => setZoomPercent(value)}
        defaultValue={0}
        orientation="vertical"
        aria-label="zoom"
      />
      <svg width={width} height={height} ref={svgRef} />
    </div>
  );
}

export default styled(DirectedGraph)`
  overflow: hidden;

  text {
    font-weight: 300;
    font-family: "Helvetica Neue", Helvetica, Arial, sans-serif;
    font-size: 12px;
  }

  .edgePath path {
    stroke: #333;
    stroke-width: 1.5px;
  }

  foreignObject {
    overflow: visible;
  }
`;
