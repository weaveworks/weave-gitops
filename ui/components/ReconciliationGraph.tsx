import { Slider } from "@mui/material";
import * as d3 from "d3";
import * as React from "react";
import styled from "styled-components";
import { useGetInventory } from "../hooks/inventory";
import { FluxObject } from "../lib/objects";
import { ReconciledObjectsAutomation } from "./AutomationDetail";
import DirectedGraph from "./DirectedGraph";
import Flex from "./Flex";
import RequestStateHandler from "./RequestStateHandler";
import Spacer from "./Spacer";

// This is working around a parcel build error.
// https://github.com/parcel-bundler/parcel/issues/8792
// https://github.com/weaveworks/weave-gitops/issues/3672
// Theory: this is tricking parcel into correctly importing d3. :shrug:
// TODO: Investigate if this can be removed safely via a component upgrade path
// eslint-disable-next-line @typescript-eslint/no-unused-expressions
d3;

interface Props {
  className?: string;
  reconciledObjectsAutomation: ReconciledObjectsAutomation;
}

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

function ReconciliationGraph({
  className,
  reconciledObjectsAutomation,
}: Props) {
  const { type, name, clusterName, namespace } = reconciledObjectsAutomation;
  const { data, isLoading, error } = useGetInventory(
    type,
    name,
    clusterName,
    namespace,
    true,
  );

  return (
    <RequestStateHandler loading={isLoading} error={error}>
      {data?.objects && (
        <Graph
          className={className}
          reconciledObjectsAutomation={reconciledObjectsAutomation}
          objects={data.objects}
        />
      )}
    </RequestStateHandler>
  );
}
interface GraphProps {
  className?: string;
  reconciledObjectsAutomation: ReconciledObjectsAutomation;
  objects: FluxObject[];
}
export const Graph = ({
  className,
  reconciledObjectsAutomation,
  objects,
}: GraphProps) => {
  const { type, name, clusterName, namespace, suspended, conditions, source } =
    reconciledObjectsAutomation;

  //add extra nodes
  const secondNode = {
    name,
    namespace,
    suspended,
    conditions,
    type,
    clusterName,
    children: objects,
    isCurrentNode: true,
  };
  const rootNode = {
    ...source,
    type: source?.kind,
    clusterName,
    children: [secondNode],
  };

  //graph numbers
  const nodeSize = {
    width: 800,
    height: 300,
    verticalSeparation: 150,
    horizontalSeparation: 100,
  };

  //use d3 to create tree structure
  const root = d3.hierarchy(rootNode, (d) => d.children);
  const makeTree = d3
    .tree()
    .nodeSize([
      nodeSize.width + nodeSize.horizontalSeparation,
      nodeSize.height + nodeSize.verticalSeparation,
    ])
    .separation(() => 1);
  const tree = makeTree(root);
  const descendants = tree.descendants();

  const links = tree.links();

  //zoom
  const defaultZoomPercent = 75;
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
  return (
    <Flex className={className} wide tall>
      <GraphDiv
        onMouseDown={handleMouseDown}
        onMouseMove={handleMouseMove}
        onMouseUp={handleMouseUp}
        //ends drag event if mouse leaves svg
        onMouseLeave={handleMouseUp}
      >
        <DirectedGraph
          descendants={descendants}
          links={links}
          nodeSize={nodeSize}
          zoomPercent={zoomPercent}
          pan={pan}
        />
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
};
export default styled(ReconciliationGraph)`
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
    margin-left: 0px;
  }
  .MuiSlider-thumb {
    width: 24px;
    height: 24px;
  }
`;
