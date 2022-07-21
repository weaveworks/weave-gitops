import * as d3 from "d3";
import * as React from "react";
import styled from "styled-components";
import { useGetReconciledObjects } from "../hooks/flux";
import { Condition, ObjectRef } from "../lib/api/core/types.pb";
import { UnstructuredObjectWithChildren } from "../lib/graph";
import { removeKind } from "../lib/utils";
import DirectedGraph from "./DirectedGraph";
import { ReconciledVisualizationProps } from "./ReconciledObjectsTable";
import RequestStateHandler from "./RequestStateHandler";

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

  return (
    <RequestStateHandler loading={isLoading} error={error}>
      <div className={className} style={{ height: "100%", width: "100%" }}>
        <DirectedGraph
          descendants={descendants}
          links={links}
          nodeSize={nodeSize}
        />
      </div>
    </RequestStateHandler>
  );
}

export default styled(ReconciliationGraph)``;
