import * as d3 from "d3";
import _ from "lodash";
import * as React from "react";
import styled from "styled-components";

type Props = {
  className?: string;
  descendants?: any[];
  links?: any[];
  hierarchy?: any;
};

function DirectedGraph({ className, descendants, links, hierarchy }: Props) {
  const [nodes, setNodes] = React.useState(null);
  const [paths, setPaths] = React.useState(null);
  const gLinks = React.useRef(null);

  // const scaleX = () =>
  //   d3
  //     .scaleLinear()
  //     .domain(
  //       d3.extent((d) => {
  //         console.log(d);
  //         d.x;
  //       })
  //     )
  //     .range([20, 1000]);
  // const scaleY = (d) =>
  //   d3
  //     .scaleLinear(d)
  //     .domain(d3.extent((d) => d.y))
  //     .range([20, 1000]);

  React.useEffect(() => {
    setNodes(descendants);
    d3.select(gLinks.current)
      .selectAll("path")
      .data(links)
      .join("path")
      .attr(
        "d",
        (d) =>
          `M${d.source.x + 550} ${d.source.y + 50}H${d.target.x + 550}V${
            d.target.y + 50
          }`
      );
  }, [descendants, links]);

  return (
    <svg width={1000} height={1000}>
      <g>
        {_.map(descendants, (d) => {
          return (
            <rect
              width={100}
              height={100}
              x={d.x + 500}
              y={d.y}
              fill="white"
              strokeWidth={2}
              stroke={"#737373"}
              style={{ borderRadius: 10 }}
            ></rect>
          );
        })}
      </g>
      <g ref={gLinks} stroke={"purple"} strokeWidth={2} fill="none" />
    </svg>
  );
}

export default styled(DirectedGraph).attrs({ className: DirectedGraph.name })`
  rect {
    border-radius: 10px;
  }
`;
