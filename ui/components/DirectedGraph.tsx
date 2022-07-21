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
};

function DirectedGraph({
  className,
  descendants,
  links,
  nodeSize,
  zoomPercent,
}: Props) {
  const zoomBox = 15000 - 14000 * (zoomPercent / 100);
  const [mouseDown, setMouseDown] = React.useState<{
    x: number;
    y: number;
    active: boolean;
  }>({ x: 0, y: 0, active: false });
  const [mouseMove, setMouseMove] = React.useState<{ x: number; y: number }>({
    x: 0,
    y: 0,
  });

  React.useEffect(() => {
    // d3.select(draggable.current).call(
    //   d3
    //     .drag()
    //     .subject(draggable.current)
    //     .on("drag", (event) => {
    //       d3.event.subject.fx = d3.event.x;
    //       d3.event.subject.fy = d3.event.y;
    //     })
    // );
  }, []);
  return (
    <svg
      width="100%"
      height="100%"
      viewBox={`${mouseMove.x} ${mouseMove.y} ${zoomBox} ${zoomBox}`}
      // onMouseDown={(e) =>
      //   setMouseDown({ x: e.clientX, y: e.clientY, active: true })
      // }
      // onMouseMove={(e) => {
      //   if (mouseDown.active)
      //     setMouseMove({
      //       x: e.clientX - mouseDown.x,
      //       y: e.clientY - mouseDown.y,
      //     });
      // }}
      // onMouseUp={(e) =>
      //   setMouseDown({ x: e.clientX, y: e.clientY, active: false })
      // }
    >
      <g stroke="#7a7a7a" strokeWidth={5} fill="none">
        {_.map(links, (l, index) => {
          // l is an object with a source and target node, each with an x and y value. M tells the path where to start, H draws a straight horizontal line, and V draws a straight vertical line
          return (
            <path
              key={index}
              d={`M${l.source.x} ${l.source.y + nodeSize.verticalSeparation}H${
                l.target.x
              }V${l.target.y + nodeSize.verticalSeparation}`}
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
    </svg>
  );
}

export default styled(DirectedGraph).attrs({ className: DirectedGraph.name })``;
