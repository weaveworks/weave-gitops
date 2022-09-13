import { Automation } from "../hooks/automations";
import { FluxObjectNode, FluxObjectNodesMap, makeObjectId } from "./objects";

// getNeighborNodes returns nodes which depend on the current node
// or are dependencies of the current node.
export function getNeighborNodes(
  nodes: FluxObjectNodesMap,
  currentNode: FluxObjectNode
): FluxObjectNode[] {
  const dependencyNodes = currentNode.dependsOn
    .map((dependency) => {
      const name = dependency.name;
      let namespace = dependency.namespace;

      if (!namespace) {
        namespace = currentNode.namespace;
      }

      return nodes[makeObjectId(namespace, name)];
    })
    .filter((n) => n);

  const nodesArray: FluxObjectNode[] = Object.values(nodes);

  const dependentNodes = nodesArray.filter((node) => {
    let isDependent = false;

    for (const dependency of node.dependsOn) {
      const name = dependency.name;
      let namespace = dependency.namespace;
      if (!namespace) {
        namespace = node.namespace;
      }

      if (name === currentNode.name && namespace === currentNode.namespace) {
        isDependent = true;
        break;
      }
    }

    return isDependent;
  });

  return dependencyNodes.concat(dependentNodes);
}

// getGraphNodes returns all nodes in the current node's dependency tree, including the current node.
export function getGraphNodes(
  nodes: FluxObjectNodesMap,
  automation: Automation
): FluxObjectNode[] {
  // Find node, corresponding to the automation.
  const currentNode =
    nodes[makeObjectId(automation.namespace, automation.name)];

  if (!currentNode) {
    return [];
  }

  currentNode.isCurrentNode = true;

  // Find nodes in the current node's dependency tree.
  let graphNodes: FluxObjectNode[] = [];

  const visitedNodes: { [name: string]: boolean } = {};
  visitedNodes[currentNode.id] = true;
  let nodesToExplore: FluxObjectNode[] = [currentNode];

  while (nodesToExplore.length > 0) {
    const node = nodesToExplore.shift();

    const newNodes = getNeighborNodes(nodes, node).filter(
      (n) => !visitedNodes[n.id]
    );

    for (const n of newNodes) {
      visitedNodes[n.id] = true;
    }

    nodesToExplore = nodesToExplore.concat(newNodes);

    graphNodes = graphNodes.concat(node);
  }

  if (graphNodes.length === 1) {
    return [];
  }

  return graphNodes;
}
