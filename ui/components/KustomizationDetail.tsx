import * as React from "react";
import { useRouteMatch } from "react-router-dom";
import styled from "styled-components";
import { AppContext } from "../contexts/AppContext";
import { useSyncAutomation } from "../hooks/automations";
import { AutomationKind, Kustomization } from "../lib/api/core/types.pb";
import Alert from "./Alert";
import EventsTable from "./EventsTable";
import Flex from "./Flex";
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
  kustomization?: Kustomization;
  className?: string;
};

const TabContent = styled(Flex)`
  margin-top: ${(props) => props.theme.spacing.medium};
  width: 100%;
  height: 100%;
`;

function KustomizationDetail({ kustomization, className }: Props) {
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
    <Flex wide tall column className={className}>
      {sync.isError && (
        <Alert
          severity="error"
          message={sync.error.message}
          title="Sync Error"
        />
      )}
      <PageStatus
        conditions={kustomization?.conditions}
        suspended={kustomization?.suspended}
      />
      <InfoList
        items={[
          ["Namespace", kustomization?.namespace],
          ["Source", <SourceLink sourceRef={kustomization?.sourceRef} />],
          ["Applied Revision", kustomization?.lastAppliedRevision],
          ["Cluster", kustomization?.clusterName],
          ["Path", kustomization?.path],
          ["Interval", <Interval interval={kustomization?.interval} />],
          [
            "Last Updated At",
            <Timestamp time={""} />,
          ],
        ]}
      />
      <SyncButton onClick={handleSyncClicked} loading={sync.isLoading} />
      <TabContent>
        <SubRouterTabs rootPath={`${path}/details`}>
          <RouterTab name="Details" path={`${path}/details`}>
            <ReconciledObjectsTable
              automationKind={AutomationKind.KustomizationAutomation}
              automationName={kustomization?.name}
              namespace={kustomization?.namespace}
              kinds={kustomization?.inventory}
              clusterName={kustomization?.clusterName}
            />
          </RouterTab>
          <RouterTab name="Events" path={`${path}/events`}>
            <EventsTable
              namespace={kustomization?.namespace}
              involvedObject={{
                kind: "Kustomization",
                name: kustomization?.name,
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
  width: 100%;

  ${Alert} {
    margin-bottom: 16px;
  }
`;
