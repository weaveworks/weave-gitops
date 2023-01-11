import * as React from "react";
import styled from "styled-components";
import { useGetObject } from "../../../hooks/objects";
import { Kind } from "../../../lib/api/core/types.pb";
import { ImageUpdateAutomation } from "../../../lib/objects";
import { V2Routes } from "../../../lib/types";
import EventsTable from "../../EventsTable";
import Flex from "../../Flex";
import InfoList, { InfoField } from "../../InfoList";
import Metadata from "../../Metadata";
import Page from "../../Page";
import PageStatus from "../../PageStatus";
import SourceLink from "../../SourceLink";
import Spacer from "../../Spacer";
import SubRouterTabs, { RouterTab } from "../../SubRouterTabs";
import Text from "../../Text";
import YamlView from "../../YamlView";
import ImageUpdateAction from "./ImageUpdateDetails";

type Props = {
  className?: string;
  name: string;
  namespace: string;
  clusterName: string;
};
function getInfoList(
  data: ImageUpdateAutomation,
  clusterName: string
): InfoField[] {
  const {
    kind,
    spec: { update, git },
  } = data.obj;
  const { path } = update;
  const { commit, checkout, push } = git;

  return [
    ["Kind", kind],
    ["Namespace", data.namespace],
    [
      "Source",
      <SourceLink sourceRef={data.sourceRef} clusterName={clusterName} />,
    ],
    ["Update Path", path],
    ["Checkout Branch", checkout?.ref?.branch],
    ["Push Branch", push.branch],
    ["Author Name", commit.author.name],
    ["Author Email", commit.author.email],
    ["Commit Template", commit.messageTemplate],
  ];
}

function ImageAutomationUpdatesDetails({
  className,
  name,
  namespace,
  clusterName,
}: Props) {
  const { data, isLoading, error } = useGetObject<ImageUpdateAutomation>(
    name,
    namespace,
    Kind.ImageUpdateAutomation,
    clusterName,
    {
      refetchInterval: 50000,
    }
  );

  const rootPath = V2Routes.ImageAutomationUpdatesDetails;
  return (
    <Page error={error} loading={isLoading} className={className}>
      {!!data && (
        <>
          <Flex wide tall column className={className}>
            <Text size="large" semiBold titleHeight>
              {data.name}
            </Text>
            <Spacer margin="xs" />
            <PageStatus
              conditions={data.conditions}
              suspended={data.suspended}
            />
            <Spacer margin="xs" />
            <ImageUpdateAction
              name={data.name}
              namespace={data.namespace}
              clusterName={data.namespace}
              kind={Kind.ImageUpdateAutomation}
            />
            <Spacer margin="xs" />
          </Flex>

          <SubRouterTabs rootPath={`${rootPath}/details`}>
            <RouterTab name="Details" path={`${rootPath}/details`}>
              <>
                <InfoList items={getInfoList(data, clusterName)} />
                <Metadata
                  metadata={data.metadata}
                  labels={data.labels}
                />
              </>
            </RouterTab>
            <RouterTab name="Events" path={`${rootPath}/events`}>
              <EventsTable
                namespace={data.namespace}
                involvedObject={{
                  kind: data.type,
                  name: data.name,
                  namespace: data.namespace,
                  clusterName: data.clusterName,
                }}
              />
            </RouterTab>
            <RouterTab name="yaml" path={`${rootPath}/yaml`}>
              <YamlView
                yaml={data.yaml}
                object={{
                  kind: data.type,
                  name: data.name,
                  namespace: data.namespace,
                }}
              />
            </RouterTab>
          </SubRouterTabs>
        </>
      )}
    </Page>
  );
}

export default styled(ImageAutomationUpdatesDetails)``;
