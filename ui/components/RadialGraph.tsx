import * as React from "react";
import styled from "styled-components";

type Props = {
  className?: string;
};

function RadialGraph({ className }: Props) {
  const svgRef = React.useRef();
  return (
    <div className={className}>
      <svg ref={svgRef}></svg>
    </div>
  );
}

export default styled(RadialGraph).attrs({ className: RadialGraph.name })``;
