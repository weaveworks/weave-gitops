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
  height: number | string;
  labelType?: "html" | "text";
  labelShape: "rect" | "ellipse";
};

const SliderFlex = styled(Flex)`
  padding-top: ${(props) => props.theme.spacing.base};
  min-height: 400px;
  min-width: 60px;
  height: 100%;
  width: 5%;
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
  const graphRef = React.useRef<D3Graph>();
  const [zoomPercent, setZoomPercent] = React.useState(20);

  React.useEffect(() => {
    if (!svgRef.current) {
      return;
    }

    // https://github.com/jsdom/jsdom/issues/2531
    if (process.env.NODE_ENV === "test") {
      return;
    }

    const graph = new D3Graph(svgRef.current, {
      labelShape,
      labelType,
      scale,
      initialZoom: zoomPercent,
    });
    graph.update(nodes, edges);
    graph.render();
    graphRef.current = graph;
  }, []);

  React.useEffect(() => {
    graphRef.current.zoom(zoomPercent);
  }, [zoomPercent]);

  React.useEffect(() => {
    graphRef.current.update(nodes, edges);
    graphRef.current.render();
  }, [nodes, edges]);

  const viewBoxOffsetX = -zoomPercent * 1.25;

  return (
    <Flex wide tall className={className}>
      <svg
        viewBox={`${viewBoxOffsetX} 0 100 100`}
        preserveAspectRatio="xMidYMid meet"
        ref={svgRef}
      />

      <Flex tall>
        <SliderFlex column center align>
          <Slider
            onChange={(e, value: number) => setZoomPercent(value)}
            defaultValue={20}
            orientation="vertical"
            aria-label="zoom"
          />
          <Spacer padding="base" />
          <PercentFlex>{zoomPercent}%</PercentFlex>
        </SliderFlex>
      </Flex>
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
    width: 650px;
    height: 200px;
    overflow: visible;
  }
  .MuiSlider-vertical {
    min-height: 400px;
  }
  .MuiSlider-vertical .MuiSlider-track {
    width: 6px;
  }
  .MuiSlider-vertical .MuiSlider-rail {
    width: 6px;
  }
  .MuiSlider-vertical .MuiSlider-thumb {
    margin-left: -9px;
  }
  .MuiSlider-thumb {
    width: 24px;
    height: 24px;
  }
`;

// D3 doesn't play nicely with React, as they both manipulate the DOM.
// Since polling was added, we re-render the graph after every poll.
// Split the D3 graphic logic into a class to avoid resetting transform state on every render.
class D3Graph {
  containerEl;
  svg;
  graph;
  opts;

  constructor(element, opts) {
    const dagreD3LibRef = dagreD3;
    this.graph = new dagreD3LibRef.graphlib.Graph();
    this.opts = opts;
    this.containerEl = element;
    this.svg = d3.select(element);
    this.svg.append("g");
    this.zoom(opts.initialZoom);
  }

  zoom(zoomPercent) {
    const zoom = d3.zoom().on("zoom", (e) => {
      e.transform.k = (zoomPercent + 20) / 1000;
      this.svg.select("g").attr("transform", e.transform);
    });

    this.svg
      .call(zoom)
      .call(zoom.transform, d3.zoomIdentity.scale(zoomPercent))
      .on("wheel.zoom", null);
  }

  update(nodes, edges) {
    this.graph
      .setGraph({
        nodesep: 100,
        ranksep: 150,
        rankdir: "TB",
      })
      .setDefaultEdgeLabel(() => {
        return {};
      });

    _.each(nodes, (n) => {
      this.graph.setNode(n.id, {
        label: n.label(n.data),
        labelType: this.opts.labelType,
        shape: this.opts.labelShape,
        width: 630,
        height: 180,
        rx: 10,
        ry: 10,
        labelClass: "node-label",
      });
    });

    _.each(edges, (e) => {
      this.graph.setEdge(e.source, e.target, { arrowhead: "undirected" });
    });
  }

  render() {
    const render = new dagreD3.render();
    render(d3.select("svg g"), this.graph);
  }
}
