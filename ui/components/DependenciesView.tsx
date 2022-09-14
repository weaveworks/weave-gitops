import * as React from "react";
import styled from "styled-components";
import { Automation } from "../hooks/automations";
import { useListObjects } from "../hooks/objects";
import {
  fluxObjectKindToKind,
  FluxObjectNode,
  FluxObjectNodesMap,
} from "../lib/objects";
import { getGraphNodes } from "../lib/dependencies";
import DagGraph from "./DagGraph";
import Flex, { MessageFlex } from "./Flex";
import RequestStateHandler from "./RequestStateHandler";
import Text from "./Text";

const NoDependenciesMessage = styled(MessageFlex)`
  & {
    margin-top: ${({ theme }) => theme.spacing.xxl};
    margin-bottom: ${({ theme }) => theme.spacing.xxl};

    h2 {
      margin-top: 0;
      margin-bottom: 0;

      &:not(:first-child) {
        margin-top: ${({ theme }) => theme.spacing.medium};
      }
    }

    p {
      margin-top: ${({ theme }) => theme.spacing.medium};
      margin-bottom: 0;
      font-size: 16px;
      line-height: 20px;
    }
  }
`;

type DependenciesViewProps = {
  className?: string;
  automation?: Automation;
};

const graphNodesPlaceholder = [] as FluxObjectNode[];

function DependenciesView({ className, automation }: DependenciesViewProps) {
  const [graphNodes, setGraphNodes] = React.useState<FluxObjectNode[] | null>(
    null
  );

  const automationKind = automation?.kind;

  const {
    data,
    isLoading: isLoadingData,
    error,
  } = automation
    ? useListObjects(
        "",
        fluxObjectKindToKind(automationKind),
        automation?.clusterName
      )
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

  const Heading = Text.withComponent("h2");
  const Paragraph = Text.withComponent("p");

  return (
    <RequestStateHandler loading={isLoading} error={error}>
      {shouldShowGraph ? (
        <DagGraph className={className} nodes={graphNodes} />
      ) : (
        <Flex className={className} wide tall center>
          <NoDependenciesMessage>
            <Heading semiBold size="large">
              No Dependencies
            </Heading>
            <Paragraph>
              There are no dependencies set up for your kustomizations or helm
              releases at this time. You can set them up using the dependsOn
              field on the kustomization or helm release object.
            </Paragraph>
            <Heading semiBold size="large">
              What are dependencies for?
            </Heading>
            <Paragraph>
              Dependencies allow you to relate different kustomizations and helm
              releases, as well as specifying an order in which your resources
              should be started. For example, you can wait for a database to
              report as 'Ready' before attempting to deploy other services.
            </Paragraph>
          </NoDependenciesMessage>
        </Flex>
      )}
    </RequestStateHandler>
  );
}

export default styled(DependenciesView).attrs({
  className: DependenciesView.name,
})``;
