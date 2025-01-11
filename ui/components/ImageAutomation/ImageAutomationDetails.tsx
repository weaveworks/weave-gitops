import React from "react";
import styled from "styled-components";
import { Kind } from "../../lib/api/core/types.pb";
import { createYamlCommand } from "../../lib/utils";
import EventsTable from "../EventsTable";
import Flex from "../Flex";
import PageStatus from "../PageStatus";
import HeaderRows, { RowItem } from "../Policies/Utils/HeaderRows";
import SubRouterTabs, { RouterTab } from "../SubRouterTabs";
import SyncActions from "../Sync/SyncActions";
import YamlView from "../YamlView";
interface Props {
  className?: string;
  data: any;
  kind: Kind;
  rootPath: string;
  infoFields: RowItem[];
  children?: any;
}

const ImageAutomationDetails = ({
  className,
  data,
  kind,
  infoFields,
  children,
}: Props) => {
  const { name, namespace, clusterName, suspended, conditions, yaml } = data;

  return (
    <Flex wide tall column className={className} gap="4">
      <PageStatus conditions={conditions} suspended={suspended} />
      {kind !== Kind.ImagePolicy && (
        <SyncActions
          name={name}
          namespace={namespace}
          clusterName={clusterName}
          kind={kind}
          suspended={suspended}
          hideSyncOptions
        />
      )}

      <SubRouterTabs rootPath="details">
        <RouterTab name="Details" path="details">
          <Flex column gap="4">
            <HeaderRows items={infoFields} />
            {children}
          </Flex>
        </RouterTab>
        <RouterTab name="Events" path="events">
          <EventsTable
            namespace={namespace}
            involvedObject={{
              kind: kind,
              name: name,
              namespace: namespace,
              clusterName: clusterName,
            }}
          />
        </RouterTab>
        <RouterTab name="yaml" path="yaml">
          <YamlView
            yaml={yaml}
            header={createYamlCommand(kind, name, namespace)}
          />
        </RouterTab>
      </SubRouterTabs>
    </Flex>
  );
};

export default styled(ImageAutomationDetails).attrs({
  className: ImageAutomationDetails.name,
})`
  ${SubRouterTabs} {
    margin-top: ${(props) => props.theme.spacing.medium};
  }
`;
