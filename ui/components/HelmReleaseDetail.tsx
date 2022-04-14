import { createHashHistory } from "history";
import * as React from "react";
import styled from "styled-components";
import { AppContext } from "../contexts/AppContext";
import { useSyncAutomation } from "../hooks/automations";
import {
  AutomationKind,
  HelmRelease,
  SourceRefSourceKind,
} from "../lib/api/core/types.pb";
import Alert from "./Alert";
import EventsTable from "./EventsTable";
import Flex from "./Flex";
import HashRouterTabs, { HashRouterTab } from "./HashRouterTabs";
import Heading from "./Heading";
import InfoList from "./InfoList";
import Interval from "./Interval";
import PageStatus from "./PageStatus";
import ReconciledObjectsTable from "./ReconciledObjectsTable";
import SourceLink from "./SourceLink";
import SyncButton from "./SyncButton";

type Props = {
  name: string;
  clusterName: string;
  helmRelease?: HelmRelease;
  className?: string;
};

const Info = styled.div`
  padding-bottom: 32px;
`;

const TabContent = styled.div`
  margin-top: 52px;
`;

export default function HelmReleaseDetail({
  name,
  helmRelease,
  className,
}: Props) {
  const { notifySuccess } = React.useContext(AppContext);
  const hashHistory = createHashHistory();
  const sync = useSyncAutomation({
    name: helmRelease?.name,
    namespace: helmRelease?.namespace,
    clusterName: helmRelease?.clusterName,
    kind: AutomationKind.HelmReleaseAutomation,
  });

  const handleSyncClicked = (opts) => {
    sync.mutateAsync(opts).then(() => {
      notifySuccess("Resource synced successfully");
    });
  };

  return (
    <Flex className={className} wide tall column align>
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
      <Flex wide>
        {sync.isError && (
          <Alert
            severity="error"
            message={sync.error.message}
            title="Sync Error"
          />
        )}
      </Flex>
      <SyncButton onClick={handleSyncClicked} loading={sync.isLoading} />
      <TabContent>
        <HashRouterTabs history={hashHistory} defaultPath="/details">
          <HashRouterTab name="Details" path="/details">
            <ReconciledObjectsTable
              kinds={helmRelease?.inventory}
              automationName={helmRelease?.name}
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
    </Flex>
  );
}
