import * as React from "react";
import styled from "styled-components";
import { useListObjects } from "../hooks/objects";
import { Kind } from "../lib/api/core/types.pb";
import { getGraphNodes } from "../lib/dependencies";
import { Automation, FluxObjectNode, FluxObjectNodesMap } from "../lib/objects";
import DagGraph from "./DagGraph";
import Flex from "./Flex";
import Link from "./Link";
import MessageBox from "./MessageBox";
import RequestStateHandler from "./RequestStateHandler";
import Spacer from "./Spacer";
import Text from "./Text";

type DependenciesViewProps = {
  className?: string;
  automation?: Automation;
};

const graphNodesPlaceholder = [] as FluxObjectNode[];

function DependenciesView({ className, automation }: DependenciesViewProps) {
  const [graphNodes, setGraphNodes] = React.useState<FluxObjectNode[] | null>(
    null,
  );

  const automationKind = Kind[automation?.type];

  const {
    data,
    isLoading: isLoadingData,
    error,
  } = automation
    ? useListObjects("", automationKind, automation?.clusterName, {})
    : { data: { objects: [], errors: [] }, error: null, isLoading: false };

  React.useEffect(() => {
    if (isLoadingData) {
      return;
    }

    if (error || data?.errors?.length > 0) {
      setGraphNodes(graphNodesPlaceholder);
      return;
    }

    const allNodes: FluxObjectNodesMap = {};
    data.objects.forEach((obj) => {
      const n = new FluxObjectNode(obj);
      allNodes[n.id] = n;
    });

    const nodes = getGraphNodes(allNodes, automation);

    nodes.sort((a, b) => a.id.localeCompare(b.id));

    if (nodes.length === 0) {
      setGraphNodes(graphNodesPlaceholder);
    } else {
      setGraphNodes(nodes);
    }
  }, [isLoadingData, data, error]);

  const isLoading = isLoadingData && !graphNodes;

  const shouldShowGraph = !!graphNodes && graphNodes.length > 0;

  return (
    <RequestStateHandler loading={isLoading} error={error}>
      {shouldShowGraph ? (
        <DagGraph className={className} nodes={graphNodes} />
      ) : (
        <Flex className={className} wide tall column align>
          <Spacer padding="xl" />
          <MessageBox gap="12">
            <Text semiBold size="large">
              No Dependencies
            </Text>
            <Text>
              Currently no dependencies are set up for your Kustomizations or
              HelmReleases. You can set them up using the dependsOn field on the{" "}
              <Link
                newTab
                href="https://fluxcd.io/flux/components/kustomize/kustomizations/#dependencies"
              >
                Kustomization
              </Link>{" "}
              or{" "}
              <Link
                newTab
                href="https://fluxcd.io/flux/components/helm/helmreleases/#helmrelease-dependencies"
              >
                Helm Release
              </Link>{" "}
              object.
            </Text>
            <Text semiBold size="large">
              What are dependencies for?
            </Text>
            <Text>
              Dependencies allow you to relate different Kustomizations and Helm
              Releases and to specify an order in which your resources should be
              started. For example, you can wait for a database to report as
              &#39;Ready&#39; before attempting to deploy other services.
            </Text>
          </MessageBox>
          <Spacer padding="xl" />
        </Flex>
      )}
    </RequestStateHandler>
  );
}

export default styled(DependenciesView).attrs({
  className: DependenciesView.name,
})``;
