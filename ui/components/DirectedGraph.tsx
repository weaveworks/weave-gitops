import Slider from "@material-ui/core/Slider";
import * as d3 from "d3";
import dagreD3 from "dagre-d3";
import _ from "lodash";
import * as React from "react";
import styled from "styled-components";
import { muiTheme } from "../lib/theme";
import Flex from "./Flex";
import Spacer from "./Spacer";

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

const SliderFlex = styled(Flex)`
  position: relative;
  min-height: 200px;
  height: 15vh;
  width: 5%;
  top: 150px;
`;

const PercentFlex = styled(Flex)`
  color: ${muiTheme.palette.primary.main};
  padding: 10px;
  background: rgba(0, 179, 236, 0.1);
  border-radius: 2px;
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

  const [zoomPercent, setZoomPercent] = React.useState(30);

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
      e.transform.k = (zoomPercent + 30) / 100;
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
    <Flex className={className}>
      <svg width={width} height={height} ref={svgRef} />
      <SliderFlex column align>
        <Slider
          onChange={(e, value: number) => setZoomPercent(value)}
          defaultValue={30}
          orientation="vertical"
          aria-label="zoom"
        />
        <Spacer padding="base" />
        <PercentFlex>{zoomPercent}%</PercentFlex>
      </SliderFlex>
    </Flex>
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
    display: flex;
    flex-direction: column;
    width: 125px;
    height: 125px;
    overflow: visible;
  }
`;
