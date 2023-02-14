import React from "react";
import { Kind } from "../../lib/api/core/types.pb";
import EventsTable from "../EventsTable";
import Flex from "../Flex";
import InfoList, { InfoField } from "../InfoList";
import PageStatus from "../PageStatus";
import Spacer from "../Spacer";
import SubRouterTabs, { RouterTab } from "../SubRouterTabs";
import Text from "../Text";
import YamlView from "../YamlView";

interface Props {
  data: any;
  kind: Kind;
  rootPath: string;
  infoFields: InfoField[];
  children?: any;
}

const ImageAutomationDetails = ({
  data,
  kind,
  infoFields,
  children,
}: Props) => {
  const { name, namespace, clusterName, suspended, conditions } = data;
  const tabs = [
    {
      name: "Details",
      path: "details",
      component: () => {
        return (
          <>
            <InfoList items={infoFields} />
            <Spacer margin="xs" />
            {children}
          </>
        );
      },
    },
    {
      name: "Events",
      path: "events",
      component: () => {
        return (
          <EventsTable
            namespace={namespace}
            involvedObject={{
              kind: kind,
              name: name,
              namespace: namespace,
              clusterName: clusterName,
            }}
          />
        );
      },
    },
    {
      name: "yaml",
      path: "yaml",
      component: () => {
        return (
          <YamlView
            yaml={data.yaml}
            object={{
              kind: kind,
              name: name,
              namespace: namespace,
            }}
          />
        );
      },
    },
  ];

  return (
    <Flex wide tall column>
      <Text size="large" semiBold titleHeight>
        {name}
      </Text>
      <Spacer margin="xs" />
      <PageStatus conditions={conditions} suspended={suspended} />
      <Spacer margin="xs" />
      {/* ImageUpdateAutomation sync is not supported yet and it'll be added in future PR */}
      {/* <SyncActions
          name={name}
          namespace={namespace}
          clusterName={clusterName}
          kind={kind}
        />
        <Spacer margin="xs" /> */}

      <SubRouterTabs tabs={tabs} />
    </Flex>
  );
};

export default ImageAutomationDetails;
