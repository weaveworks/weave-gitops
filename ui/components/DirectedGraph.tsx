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
        d3
          .linkHorizontal()
          .source((d) => [d.x, d.y])
          .target((d) => [d.x, d.y])
      );
  }, [descendants, links]);

  return (
    <svg width={1000} height={1000}>
      <g>
        {_.map(descendants, (d) => {
          return <rect width={100} height={100} x={d.x} y={d.y}></rect>;
        })}
      </g>
      <g ref={gLinks} />
    </svg>
  );
}

export default styled(DirectedGraph).attrs({ className: DirectedGraph.name })``;
