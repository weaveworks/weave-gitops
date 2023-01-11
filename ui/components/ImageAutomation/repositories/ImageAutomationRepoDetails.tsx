import * as React from "react";
import styled from "styled-components";
import Page from "../../Page";
import { useGetObject } from "../../../hooks/objects";
import { Kind } from "../../../lib/api/core/types.pb";
import { ImageRepository, ImageUpdateAutomation } from "../../../lib/objects";
import { V2Routes } from "../../../lib/types";
import EventsTable from "../../EventsTable";
import Flex from "../../Flex";
import InfoList from "../../InfoList";
import Interval from "../../Interval";
import PageStatus from "../../PageStatus";
import SourceLink from "../../SourceLink";
import Spacer from "../../Spacer";
import SubRouterTabs, { RouterTab } from "../../SubRouterTabs";
import Text from "../../Text";
import YamlView from "../../YamlView";

type Props = {
  className?: string;
  name: string;
  namespace: string;
  clusterName: string;
};

function ImageAutomationRepoDetails({
  className,
  name,
  namespace,
  clusterName,
}: Props) {
  const { data, isLoading, error } = useGetObject<ImageRepository>(
    name,
    namespace,
    Kind.ImageRepository,
    clusterName,
    {
      refetchInterval: 50000,
    }
  );
  const rootPath = V2Routes.ImageAutomationRepositoriesDetails;
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

            <InfoList
              items={[
                [
                  "Source",
                  <SourceLink
                    sourceRef={data.sourceRef}
                    clusterName={clusterName}
                  />,
                ],
                ["Namespace", data.namespace],
                ["Interval", <Interval interval={data.interval} />],
                ["Tag Count", data.tagCount],
              ]}
            />
          </Flex>
          <Spacer margin="xs" />

          <SubRouterTabs rootPath={`${rootPath}/policies`}>
            <RouterTab name="Policies" path={`${rootPath}/policies`}>
              <p>Available policies</p>
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

export default styled(ImageAutomationRepoDetails)``;
