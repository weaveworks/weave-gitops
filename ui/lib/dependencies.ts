import { Automation } from "../hooks/automations";
import { FluxObjectNode } from "./objects";

// findNode searches for a node by the node's name and namespace.
export function findNode(
  nodes: FluxObjectNode[],
  name: string,
  namespace: string
): FluxObjectNode | null {
  const matches = nodes.filter(
    (node) => node.name === name && node.namespace === namespace
  );

  if (matches.length > 0) {
    return matches[0];
  } else {
    return null;
  }
}

// getNeighborNodes returns nodes which depend on the current node
// or are dependencies of the current node.
export function getNeighborNodes(
  nodes: FluxObjectNode[],
  currentNode: FluxObjectNode
): FluxObjectNode[] {
  const dependencyNodes = currentNode.dependsOn
    .map((dependency) => {
      const name = dependency.name;
      let namespace = dependency.namespace;

      if (!namespace) {
        namespace = currentNode.namespace;
      }

      return findNode(nodes, name, namespace);
    })
    .filter((n) => n);

  const dependentNodes = nodes.filter((node) => {
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
  nodes: FluxObjectNode[],
  automation: Automation
): FluxObjectNode[] {
  // Find node, corresponding to the automation.
  const currentNode = findNode(nodes, automation.name, automation.namespace);

  if (!currentNode) {
    return [];
  }

  currentNode.isCurrentNode = true;

  // Find nodes in the current node's dependency tree.
  let graphNodes: FluxObjectNode[] = [];

  const visitedNodes = new Map<string, boolean>();

  let nodesToExplore: FluxObjectNode[] = [currentNode].concat(
    getNeighborNodes(nodes, currentNode)
  );

  if (nodesToExplore.length === 1) {
    return [];
  }

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

  // Remove duplicates from graphNodes.
  graphNodes = graphNodes.filter(
    (node, index) => graphNodes.indexOf(node) === index
  );

  return graphNodes;
}
