import _ from "lodash";
import * as React from "react";
import type { JSX } from "react";
import styled from "styled-components";
import { useListAutomations } from "../hooks/automations";
import { Kind } from "../lib/api/core/types.pb";
import { HelmRelease, OCIRepository, Source } from "../lib/objects";
import { createYamlCommand, getSourceRefForAutomation } from "../lib/utils";
import AutomationsTable from "./AutomationsTable";
import EventsTable from "./EventsTable";
import Flex from "./Flex";
import InfoList, { InfoField } from "./InfoList";
import LoadingPage from "./LoadingPage";
import Metadata from "./Metadata";
import PageStatus from "./PageStatus";
import SubRouterTabs, { RouterTab } from "./SubRouterTabs";
import SyncActions from "./Sync/SyncActions";
import YamlView from "./YamlView";

//must specify OCIRepository type, artifactMetadata causes errors on the Source type
const getArtifactMetadata = (source: OCIRepository) => {
  return source.artifactMetadata || null;
};

type Props = {
  className?: string;
  type: Kind;
  children?: JSX.Element;
  source: Source;
  info: InfoField[];
  customActions?: JSX.Element[];
};

function SourceDetail({ className, source, info, type, customActions }: Props) {
  const {
    name,
    namespace,
    yaml,
    clusterName,
    conditions,
    suspended,
    metadata,
    labels,
  } = source;

  const { data: automations, isLoading: automationsLoading } =
    useListAutomations();

  if (automationsLoading) {
    return <LoadingPage />;
  }

  const isNameRelevant = (expectedName) => {
    return expectedName == name;
  };

  const isRelevant = (expectedType, expectedName) => {
    return expectedType == type && isNameRelevant(expectedName);
  };

  const relevantAutomations = _.filter(automations?.result, (a) => {
    if (!source) {
      return false;
    }
    if (a.clusterName != clusterName) {
      return false;
    }

    if (type === Kind.HelmChart) {
      return isNameRelevant((a as HelmRelease)?.helmChart?.name);
    }

    const sourceRef = getSourceRefForAutomation(a);

    return isRelevant(sourceRef?.kind, sourceRef?.name);
  });

  return (
    <Flex wide tall column className={className} gap="32">
      <Flex column gap="8">
        <PageStatus conditions={conditions} suspended={suspended} />
        <SyncActions
          name={name}
          namespace={namespace}
          clusterName={clusterName}
          kind={type}
          suspended={suspended}
          hideSyncOptions
          customActions={customActions}
        />
      </Flex>

      <SubRouterTabs rootPath="details">
        <RouterTab name="Details" path="details">
          <>
            <InfoList items={info} />
            <Metadata
              metadata={metadata}
              labels={labels}
              artifactMetadata={
                type === Kind.OCIRepository &&
                getArtifactMetadata(source as OCIRepository)
              }
            />
            <AutomationsTable automations={relevantAutomations} hideSource />
          </>
        </RouterTab>
        <RouterTab name="Events" path="events">
          <EventsTable
            namespace={namespace}
            involvedObject={{
              kind: type,
              name: name,
              namespace: namespace,
              clusterName: clusterName,
            }}
          />
        </RouterTab>
        <RouterTab name="yaml" path="yaml">
          <YamlView
            yaml={yaml}
            header={createYamlCommand(type, name, namespace)}
          />
        </RouterTab>
      </SubRouterTabs>
    </Flex>
  );
}

export default styled(SourceDetail).attrs({ className: SourceDetail.name })`
  ${SubRouterTabs} {
    margin-top: ${(props) => props.theme.spacing.medium};
  }
`;
