import { createHashHistory } from "history";
import * as React from "react";
import styled from "styled-components";
import {
  HelmRelease,
  AutomationKind,
  SourceRefSourceKind,
} from "../lib/api/core/types.pb";
import { WeGONamespace } from "../lib/types";
import EventsTable from "./EventsTable";
import Flex from "./Flex";
import HashRouterTabs, { HashRouterTab } from "./HashRouterTabs";
import Heading from "./Heading";
import InfoList from "./InfoList";
import Interval from "./Interval";
import PageStatus from "./PageStatus";
import ReconciledObjectsTable from "./ReconciledObjectsTable";
import SourceLink from "./SourceLink";
import Spacer from "./Spacer";

type Props = {
  name: string;
  clusterName: string;
  helmRelease?: HelmRelease;
}

const Info = styled.div`
  padding-bottom: 32px;
`;

const TabContent = styled.div`
  margin-top: 52px;
`;

export default function HelmReleaseDetail({ name, helmRelease }: Props) {
  const hashHistory = createHashHistory();
  return (
    <>
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
              clusterName={helmRelease?.clusterName}
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
    </>
  );
}
