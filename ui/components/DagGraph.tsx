import _ from "lodash";
import * as React from "react";
import * as d3d from "d3-dag";
import styled from "styled-components";
import { Slider } from "@material-ui/core";
import { FluxObjectNode } from "../lib/objects";
import Flex from "./Flex";
import DagGraphNode from "./DagGraphNode";
import Spacer from "./Spacer";

type NodeSize = {
  width: number;
  height: number;
  verticalSeparation: number;
  horizontalSeparation: number;
  arrowWidth: number;
  arrowHeight: number;
};

type Props = {
  className?: string;
  nodes: FluxObjectNode[];
};

const SliderFlex = styled(Flex)`
  padding-top: ${(props) => props.theme.spacing.base};
  min-height: 400px;
  min-width: 60px;
  width: 5%;
`;

const PercentFlex = styled(Flex)`
  color: ${(props) => props.theme.colors.primary10};
  padding: 10px;
  background: rgba(0, 179, 236, 0.1);
  border-radius: 2px;
`;

const GraphDiv = styled.div`
  width: 100%;
  height: 100%;
`;

function DagGraph({ className, nodes }: Props) {
  //zoom
  const defaultZoomPercent = 85;
  const [zoomPercent, setZoomPercent] = React.useState(defaultZoomPercent);

  //pan
  const [pan, setPan] = React.useState({ x: 0, y: 0 });
  const [isPanning, setIsPanning] = React.useState(false);
  const handleMouseDown = () => {
    setIsPanning(true);
  };
  const handleMouseMove = (e) => {
    //viewBox change. e.movement is change since previous mouse event
    if (isPanning) setPan({ x: pan.x + e.movementX, y: pan.y + e.movementY });
  };
  const handleMouseUp = () => {
    setIsPanning(false);
  };

  //minimum zoomBox is 1000
  const zoomBox = 15000 - 14000 * (zoomPercent / 100);
  //since viewbox is so large, make smaller mouse movements correspond to larger pan
  const svgRef = React.useRef<SVGSVGElement>();
  let panScale = 1;
  if (svgRef.current) {
    panScale = zoomBox / svgRef.current.getBoundingClientRect().width;
  }

  //graph numbers
  const nodeSize: NodeSize = {
    width: 800,
    height: 300,
    verticalSeparation: 150,
    horizontalSeparation: 100,
    arrowWidth: 25,
    arrowHeight: 25,
  };

  const arrowHalfWidth = nodeSize.arrowWidth / 2;
  const nodeHalfHeight = nodeSize.height / 2;
  const linkStrokeWidth = 5;

  //use d3 to create DAG structure
  const stratify = d3d.dagStratify();
  const root = stratify(nodes);
  const makeDag = d3d
    .sugiyama()
    .nodeSize(() => [
      nodeSize.width + nodeSize.horizontalSeparation,
      nodeSize.height + nodeSize.verticalSeparation,
    ]);
  const { width } = makeDag(root);
  const descendants = root.descendants();
  const links = root.links();

  const offsetX = zoomBox / 2 - width / 2;

  return (
    <Flex className={className} wide tall>
      <GraphDiv
        onMouseDown={handleMouseDown}
        onMouseMove={handleMouseMove}
        onMouseUp={handleMouseUp}
        //ends drag event if mouse leaves svg
        onMouseLeave={handleMouseUp}
      >
        <svg
          className={className}
          width="100%"
          height="100%"
          viewBox={`${-pan.x * panScale} ${
            -pan.y * panScale
          } ${zoomBox} ${zoomBox}`}
          ref={svgRef}
        >
          <g transform={`translate(${offsetX}, 50)`}>
            <g stroke="#7a7a7a" strokeWidth={linkStrokeWidth} fill="none">
              {_.map(links, (l, index) => {
                const verticalHalf = (l.target.y - l.source.y) / 2;
                // l is an object with a source and target node, each with an x and y value.
                // M tells the path where to start,
                // H draws a straight horizontal line (capital letter means absolute coordinates),
                // and v draws a straight vertical line (lowercase letter means relative to current position)
                return (
                  <g key={index}>
                    <path
                      fill="#7a7a7a"
                      d={`M${l.target.x}, ${l.target.y}
                      l${arrowHalfWidth}, ${-nodeSize.arrowHeight}
                      h${-nodeSize.arrowWidth}
                      l${arrowHalfWidth}, ${nodeSize.arrowHeight}`}
                    />
                    <path
                      d={`M${l.source.x}, ${
                        l.source.y + nodeSize.verticalSeparation
                      }
                      v${verticalHalf}
                      H${l.target.x}
                      v${verticalHalf - nodeHalfHeight}`}
                    />
                  </g>
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
                    // fill="white"
                    strokeWidth={2}
                    stroke={"#737373"}
                    overflow="visible"
                  >
                    <DagGraphNode
                      object={d.data}
                      isCurrentNode={d.data.isCurrentNode}
                    />
                  </foreignObject>
                );
              })}
            </g>
          </g>
        </svg>
      </GraphDiv>
      <SliderFlex tall column align>
        <Slider
          onChange={(_, value: number) => setZoomPercent(value)}
          defaultValue={defaultZoomPercent}
          orientation="vertical"
          aria-label="zoom"
          min={5}
        />
        <Spacer padding="xs" />
        <PercentFlex>{zoomPercent}%</PercentFlex>
      </SliderFlex>
    </Flex>
  );
}

export default styled(DagGraph).attrs({ className: DagGraph.name })`
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
