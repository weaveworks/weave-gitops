import { createHashHistory } from "history";
import * as React from "react";
import styled from "styled-components";
import EventsTable from "../../components/EventsTable";
import Flex from "../../components/Flex";
import HashRouterTabs, { HashRouterTab } from "../../components/HashRouterTabs";
import Heading from "../../components/Heading";
import InfoList from "../../components/InfoList";
import Interval from "../../components/Interval";
import Page from "../../components/Page";
import PageStatus from "../../components/PageStatus";
import ReconciledObjectsTable from "../../components/ReconciledObjectsTable";
import SourceLink from "../../components/SourceLink";
import Spacer from "../../components/Spacer";
import { useGetHelmRelease } from "../../hooks/automations";
import {
  AutomationKind,
  SourceRefSourceKind,
} from "../../lib/api/core/types.pb";
import { WeGONamespace } from "../../lib/types";

type Props = {
  name: string;
  clusterName: string;
  className?: string;
};

const Info = styled.div`
  padding-bottom: 32px;
`;

const TabContent = styled.div`
  margin-top: 52px;
`;

function HelmReleaseDetail({ className, name, clusterName }: Props) {
  const { data, isLoading, error } = useGetHelmRelease(name, clusterName);
  const helmRelease = data?.helmRelease;
  const hashHistory = createHashHistory();

  return (
    <Page loading={isLoading} error={error} className={className} title={name}>
      <Spacer padding="xs" />
      <Flex wide between>
        <Info>
          <Heading level={2}>{helmRelease?.namespace}</Heading>
          <InfoList
            items={[
              [
                "Source",
                <SourceLink
                  sourceRef={{
                    kind: SourceRefSourceKind.HelmChart,
                    name: helmRelease?.helmChart.chart,
                  }}
                />,
              ],
              ["Chart", helmRelease?.helmChart.chart],
              ["Cluster", helmRelease?.clusterName],
              ["Interval", <Interval interval={helmRelease?.interval} />],
            ]}
          />
        </Info>
        <PageStatus
          conditions={helmRelease?.conditions}
          suspended={helmRelease?.suspended}
        />
      </Flex>
      <TabContent>
        <HashRouterTabs history={hashHistory} defaultPath="/details">
          <HashRouterTab name="Details" path="/details">
            <ReconciledObjectsTable
              kinds={helmRelease?.inventory}
              automationName={helmRelease?.name}
              namespace={WeGONamespace}
              automationKind={AutomationKind.HelmReleaseAutomation}
            />
          </HashRouterTab>
          <HashRouterTab name="Events" path="/events">
            <EventsTable
              involvedObject={{
                kind: "HelmRelease",
                name,
                namespace: helmRelease?.namespace,
              }}
            />
          </HashRouterTab>
        </HashRouterTabs>
      </TabContent>
    </Page>
  );
}

export default styled(HelmReleaseDetail).attrs({
  className: HelmReleaseDetail.name,
})``;
