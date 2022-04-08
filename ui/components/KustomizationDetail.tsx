import * as React from "react";
import { useRouteMatch } from "react-router-dom";
import styled from "styled-components";
import { AppContext } from "../contexts/AppContext";
import { useSyncAutomation } from "../hooks/automations";
import { AutomationKind, Kustomization } from "../lib/api/core/types.pb";
import { WeGONamespace } from "../lib/types";
import Alert from "./Alert";
import EventsTable from "./EventsTable";
import Flex from "./Flex";
import Heading from "./Heading";
import InfoList from "./InfoList";
import Interval from "./Interval";
import PageStatus from "./PageStatus";
import ReconciledObjectsTable from "./ReconciledObjectsTable";
import ReconciliationGraph from "./ReconciliationGraph";
import SourceLink from "./SourceLink";
import SubRouterTabs, { RouterTab } from "./SubRouterTabs";
import SyncButton from "./SyncButton";
import Timestamp from "./Timestamp";

type Props = {
  name: string;
  kustomization?: Kustomization;
  className?: string;
};

const Info = styled.div`
  margin-bottom: 16px;
`;

const TabContent = styled(Flex)`
  margin-top: 52px;
  width: 100%;
  height: 100%;
`;

function KustomizationDetail({ kustomization, name, className }: Props) {
  const { notifySuccess } = React.useContext(AppContext);
  const { path } = useRouteMatch();

  const sync = useSyncAutomation({
    name: kustomization?.name,
    namespace: kustomization?.namespace,
    clusterName: kustomization?.clusterName,
    kind: AutomationKind.KustomizationAutomation,
  });

  const handleSyncClicked = (opts) => {
    sync.mutateAsync(opts).then(() => {
      notifySuccess("Resource synced successfully");
    });
  };

  return (
    <Flex wide tall column align className={className}>
      <Flex wide between>
        <Info>
          <Heading level={2}>{kustomization?.namespace}</Heading>
          <InfoList
            items={[
              ["Source", <SourceLink sourceRef={kustomization?.sourceRef} />],
              ["Applied Revision", kustomization?.lastAppliedRevision],
              ["Cluster", kustomization?.clusterName],
              ["Path", kustomization?.path],
              ["Interval", <Interval interval={kustomization?.interval} />],
              [
                "Last Updated At",
                <Timestamp time={kustomization?.lastHandledReconciledAt} />,
              ],
            ]}
          />
        </Info>
        <PageStatus
          conditions={kustomization?.conditions}
          suspended={kustomization?.suspended}
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
        <SubRouterTabs rootPath={`${path}/details`}>
          <RouterTab name="Details" path={`${path}/details`}>
            <ReconciledObjectsTable
              automationKind={AutomationKind.KustomizationAutomation}
              automationName={kustomization?.name}
              kinds={kustomization?.inventory}
              namespace={WeGONamespace}
              clusterName={kustomization?.clusterName}
            />
          </RouterTab>
          <RouterTab name="Events" path={`${path}/events`}>
            <EventsTable
              involvedObject={{
                kind: "Kustomization",
                name,
                namespace: kustomization?.namespace,
              }}
            />
          </RouterTab>
          <RouterTab name="Graph" path={`${path}/graph`}>
            <ReconciliationGraph
              automationKind={AutomationKind.KustomizationAutomation}
              automationName={kustomization?.name}
              kinds={kustomization?.inventory}
              parentObject={kustomization}
              namespace={WeGONamespace}
              clusterName={kustomization?.clusterName}
            />
          </RouterTab>
        </SubRouterTabs>
      </TabContent>
    </Flex>
  );
}

export default styled(KustomizationDetail).attrs({
  className: KustomizationDetail.name,
})`
  ${Alert} {
    margin-bottom: 16px;
  }
`;
