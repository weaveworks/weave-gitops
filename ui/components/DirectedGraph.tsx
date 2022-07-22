import _ from "lodash";
import * as React from "react";
import styled from "styled-components";
import GraphNode from "./GraphNode";

type nodeSize = {
  width: number;
  height: number;
  verticalSeparation: number;
  horizontalSeparation: number;
};

type Props = {
  className?: string;
  descendants: any[];
  links: any[];
  nodeSize: nodeSize;
  zoomPercent: number;
  pan: { x: number; y: number };
};

function DirectedGraph({
  className,
  descendants,
  links,
  nodeSize,
  zoomPercent,
  pan,
}: Props) {
  //minimum zoomBox is 1000
  const zoomBox = 15000 - 14000 * (zoomPercent / 100);
  //since viewbox is so large, make smaller mouse movements correspond to larger pan
  const panScale = 300 / zoomPercent;

  return (
    <svg
      className={className}
      width="100%"
      height="100%"
      viewBox={`${-pan.x * panScale} ${
        -pan.y * panScale
      } ${zoomBox} ${zoomBox}`}
    >
      <g transform={`translate(${zoomBox / 2}, 50)`}>
        <g stroke="#7a7a7a" strokeWidth={5} fill="none">
          {_.map(links, (l, index) => {
            // l is an object with a source and target node, each with an x and y value. M tells the path where to start, H draws a straight horizontal line, and V draws a straight vertical line
            return (
              <path
                key={index}
                d={`M${l.source.x} ${
                  l.source.y + nodeSize.verticalSeparation
                }H${l.target.x}V${l.target.y + nodeSize.verticalSeparation}`}
              />
            );
          })}
        </g>
        <g>
          {_.map(descendants, (d, index) => {
            //turn each descendant into a GraphNode
            return (
              <foreignObject
                width={nodeSize.width}
                height={nodeSize.height}
                key={index}
                transform={`translate(${d.x - nodeSize.width / 2}, ${d.y})`}
                fill="white"
                strokeWidth={2}
                stroke={"#737373"}
                overflow="visible"
              >
                <GraphNode object={d.data} />
              </foreignObject>
            );
          })}
        </g>
      </g>
    </svg>
  );
}

export default styled(DirectedGraph).attrs({ className: DirectedGraph.name })``;
