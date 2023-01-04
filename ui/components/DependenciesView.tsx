import * as React from "react";
import styled from "styled-components";
import { useListObjects } from "../hooks/objects";
import { Kind } from "../lib/api/core/types.pb";
import { Automation } from "../lib/objects";
import DependenciesGraph from "./DependenciesGraph";
import Flex from "./Flex";
import MessageBox from "./MessageBox";
import RequestStateHandler from "./RequestStateHandler";
import Spacer from "./Spacer";
import Text from "./Text";

const NoDependenciesMessage = styled(MessageBox)`
  & {
    h2 {
      margin-top: 0;
      margin-bottom: 0;
    }

    p {
      margin-top: 0;
      margin-bottom: 0;
      font-size: 16px;
      line-height: 20px;
    }
  }
`;

const Heading = Text.withComponent("h2");
const Paragraph = Text.withComponent("p");

type DependenciesViewProps = {
  className?: string;
  automation?: Automation;
};

function DependenciesView({ className, automation }: DependenciesViewProps) {
  const automationKind = Kind[automation?.type];

  const { data, isLoading, error } = automation
    ? useListObjects("", automationKind, automation?.clusterName, {})
    : { data: { objects: [], errors: [] }, error: null, isLoading: false };

  const shouldShowGraph = !!data.objects && data.objects.length > 0;

  return (
    <RequestStateHandler loading={isLoading} error={error}>
      {shouldShowGraph ? (
        <DependenciesGraph
          data={data.objects}
          name={automation?.name}
          namespace={automation?.namespace}
        />
      ) : (
        <Flex className={className} wide tall column align>
          <Spacer padding="xl" />
          <NoDependenciesMessage>
            <Heading semiBold size="large">
              No Dependencies
            </Heading>
            <Spacer padding="xs" />
            <Paragraph>
              There are no dependencies set up for your kustomizations or helm
              releases at this time. You can set them up using the dependsOn
              field on the kustomization or helm release object.
            </Paragraph>
            <Spacer padding="xs" />
            <Heading semiBold size="large">
              What are dependencies for?
            </Heading>
            <Spacer padding="xs" />
            <Paragraph>
              Dependencies allow you to relate different kustomizations and helm
              releases, as well as specifying an order in which your resources
              should be started. For example, you can wait for a database to
              report as 'Ready' before attempting to deploy other services.
            </Paragraph>
          </NoDependenciesMessage>
          <Spacer padding="xl" />
        </Flex>
      )}
    </RequestStateHandler>
  );
}

export default styled(DependenciesView).attrs({
  className: DependenciesView.name,
})``;
