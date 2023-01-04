import * as React from "react";
import styled from "styled-components";
import { getGraphNodes } from "../lib/dependencies";
import { FluxObject, FluxObjectNode, FluxObjectNodesMap } from "../lib/objects";
import DagGraph from "./DagGraph";

type Props = {
  className?: string;
  data: FluxObject[];
  name?: string;
  namespace?: string;
};

function DependenciesGraph({ className, data, name, namespace }: Props) {
  const [graphNodes, setGraphNodes] = React.useState<FluxObjectNode[] | null>(
    null
  );
  const graphNodesPlaceholder = [] as FluxObjectNode[];

  React.useEffect(() => {
    const allNodes: FluxObjectNodesMap = {};
    data.forEach((obj) => {
      const n = new FluxObjectNode(obj);
      allNodes[n.id] = n;
    });

    const nodes = getGraphNodes(allNodes, { name, namespace });

    nodes.sort((a, b) => a.id.localeCompare(b.id));

    if (nodes.length === 0) {
      setGraphNodes(graphNodesPlaceholder);
    } else {
      setGraphNodes(nodes);
    }
  }, [data]);

  return <DagGraph className={className} nodes={graphNodes} />;
}

export default styled(DependenciesGraph).attrs({
  className: DependenciesGraph.name,
})``;
