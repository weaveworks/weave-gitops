import { Slider } from "@material-ui/core";
import * as d3 from "d3";
import * as React from "react";
import styled from "styled-components";
import { useGetReconciledObjects } from "../hooks/flux";
import { Condition, ObjectRef } from "../lib/api/core/types.pb";
import { UnstructuredObjectWithChildren } from "../lib/graph";
import { removeKind } from "../lib/utils";
import DirectedGraph from "./DirectedGraph";
import Flex from "./Flex";
import { ReconciledVisualizationProps } from "./ReconciledObjectsTable";
import RequestStateHandler from "./RequestStateHandler";
import Spacer from "./Spacer";

export type Props = ReconciledVisualizationProps & {
  parentObject: {
    name?: string;
    namespace?: string;
    conditions?: Condition[];
    suspended?: boolean;
    children?: UnstructuredObjectWithChildren[];
  };
  source: ObjectRef;
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

function ReconciliationGraph({
  className,
  parentObject,
  automationKind,
  kinds,
  clusterName,
  source,
}: Props) {
  //grab data
  const {
    data: objects,
    error,
    isLoading,
  } = parentObject
    ? useGetReconciledObjects(
        parentObject?.name,
        parentObject?.namespace,
        automationKind,
        kinds,
        clusterName
      )
    : { data: [], error: null, isLoading: false };
  //add extra nodes
  const secondNode = parentObject;
  secondNode.children = objects;
  const rootNode = {
    ...source,
    kind: removeKind(source.kind),
    children: [secondNode],
  };
  //graph numbers
  const nodeSize = {
    width: 750,
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
    ]);
  const tree = makeTree(root);
  const descendants = tree.descendants();
  const links = tree.links();

  const defaultZoomPercent = 85;
  const [zoomPercent, setZoomPercent] = React.useState(defaultZoomPercent);

  return (
    <RequestStateHandler loading={isLoading} error={error}>
      <Flex className={className} wide tall>
        <DirectedGraph
          descendants={descendants}
          links={links}
          nodeSize={nodeSize}
          zoomPercent={zoomPercent}
        />
        <SliderFlex tall column align>
          <Slider
            onChange={(_, value: number) => setZoomPercent(value)}
            defaultValue={defaultZoomPercent}
            orientation="vertical"
            aria-label="zoom"
          />
          <Spacer padding="xs" />
          <PercentFlex>{zoomPercent}%</PercentFlex>
        </SliderFlex>
      </Flex>
    </RequestStateHandler>
  );
}

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
    margin-left: -9px;
  }
  .MuiSlider-thumb {
    width: 24px;
    height: 24px;
  }
`;
