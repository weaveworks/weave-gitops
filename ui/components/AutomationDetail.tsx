import * as React from "react";
import { useRouteMatch } from "react-router-dom";
import styled from "styled-components";
import { AppContext } from "../contexts/AppContext";
import { Automation, useSyncAutomation } from "../hooks/automations";
import { AutomationKind } from "../lib/api/core/types.pb";
import Alert from "./Alert";
import EventsTable from "./EventsTable";
import Flex from "./Flex";
import Heading from "./Heading";
import InfoList, { InfoField } from "./InfoList";
import PageStatus from "./PageStatus";
import ReconciledObjectsTable from "./ReconciledObjectsTable";
import ReconciliationGraph from "./ReconciliationGraph";
import SubRouterTabs, { RouterTab } from "./SubRouterTabs";
import SyncButton from "./SyncButton";

type Props = {
  automation?: Automation;
  className?: string;
  info: InfoField[];
};

const Info = styled.div`
  margin-bottom: 16px;
`;

const TabContent = styled(Flex)`
  margin-top: 52px;
  width: 100%;
  height: 100%;
`;

function AutomationDetail({ automation, className, info }: Props) {
  const { notifySuccess } = React.useContext(AppContext);
  const { path } = useRouteMatch();

  const sync = useSyncAutomation({
    name: automation?.name,
    namespace: automation?.namespace,
    clusterName: automation?.clusterName,
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
          <Heading level={2}>{automation?.namespace}</Heading>
          <InfoList items={info} />
        </Info>
        <PageStatus
          conditions={automation?.conditions}
          suspended={automation?.suspended}
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
              automationName={automation?.name}
              namespace={automation?.namespace}
              kinds={automation?.inventory}
              clusterName={automation?.clusterName}
            />
          </RouterTab>
          <RouterTab name="Events" path={`${path}/events`}>
            <EventsTable
              namespace={automation?.namespace}
              involvedObject={{
                kind: "Kustomization",
                name: automation?.name,
                namespace: automation?.namespace,
              }}
            />
          </RouterTab>
          <RouterTab name="Graph" path={`${path}/graph`}>
            <ReconciliationGraph
              automationKind={AutomationKind.KustomizationAutomation}
              automationName={automation?.name}
              kinds={automation?.inventory}
              parentObject={automation}
              clusterName={automation?.clusterName}
            />
          </RouterTab>
        </SubRouterTabs>
      </TabContent>
    </Flex>
  );
}

export default styled(AutomationDetail).attrs({
  className: AutomationDetail.name,
})`
  width: 100%;

  ${Alert} {
    margin-bottom: 16px;
  }
`;
