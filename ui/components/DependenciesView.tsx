import * as React from "react";
import styled from "styled-components";
import { Automation } from "../hooks/automations";
import { useListObjects } from "../hooks/objects";
import { fluxObjectKindToKind, FluxObjectNode } from "../lib/objects";
import { getGraphNodes } from "../lib/dependencies";
import Flex from "./Flex";
import DagGraph from "./DagGraph";
import RequestStateHandler from "./RequestStateHandler";

type MessageProps = {
  className?: string;
};

function UnstyledMessage({ className }: MessageProps) {
  return (
    <Flex className={className} wide tall center>
      <Flex column align shadow>
        <h2>No Dependencies</h2>

        <div>
          There are no dependencies set up for your kustomizations or helm
          releases at this time. You can set them up using the dependsOn field
          on the kustomization or helm release object.
        </div>

        <h2>What are dependencies for?</h2>

        <div>
          Dependencies allow you to relate different kustomizations and helm
          releases, as well as specifying an order in which your resources
          should be started. For example, you can wait for a database to report
          as 'Ready' before attempting to deploy other services.
        </div>
      </Flex>
    </Flex>
  );
}

const Message = styled(UnstyledMessage)`
  & {
    ${Flex} {
      box-sizing: border-box;
      width: 560px;
      padding-top: ${(props) => props.theme.spacing.medium};
      padding-right: ${(props) => props.theme.spacing.xl};
      padding-bottom: ${(props) => props.theme.spacing.xxl};
      padding-left: ${(props) => props.theme.spacing.xl};
      margin-top: ${(props) => props.theme.spacing.xxl};
      margin-bottom: ${(props) => props.theme.spacing.xxl};
      border-radius: 10px;
      background-color: #ffffffd9;
      color: ${(props) => props.theme.colors.neutral30};

      h2 {
        margin-top: 0;
        margin-bottom: 0;
        font-size: ${(props) => props.theme.fontSizes.large};
        font-weight: 600;

        &:not(:first-child) {
          margin-top: ${(props) => props.theme.spacing.medium};
        }
      }

      div {
        margin-top: ${(props) => props.theme.spacing.medium};
        font-size: 16px;
        font-weight: 400;
        line-height: 20px;
      }
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

    const nodes = getGraphNodes(
      data.objects.map((obj) => new FluxObjectNode(obj)),
      automation
    );

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
      {!isLoading && (
        <>
          {shouldShowGraph ? (
            <DagGraph className={className} nodes={graphNodes} />
          ) : (
            <Message />
          )}
        </>
      )}
    </RequestStateHandler>
  );
}

export default styled(DependenciesView).attrs({
  className: DependenciesView.name,
})``;
