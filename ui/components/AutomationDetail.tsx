import * as React from "react";
import { useRouteMatch } from "react-router-dom";
import styled from "styled-components";
import { AppContext } from "../contexts/AppContext";
import { Automation, useSyncFluxObject } from "../hooks/automations";
import { useToggleSuspend } from "../hooks/flux";
import { FluxObjectKind } from "../lib/api/core/types.pb";
import Alert from "./Alert";
import Button from "./Button";
import DetailTitle from "./DetailTitle";
import EventsTable from "./EventsTable";
import Flex from "./Flex";
import InfoList, { InfoField } from "./InfoList";
import PageStatus from "./PageStatus";
import ReconciledObjectsTable from "./ReconciledObjectsTable";
import ReconciliationGraph from "./ReconciliationGraph";
import Spacer from "./Spacer";
import SubRouterTabs, { RouterTab } from "./SubRouterTabs";
import SyncButton from "./SyncButton";
type Props = {
  automation?: Automation;
  className?: string;
  info: InfoField[];
};

function AutomationDetail({ automation, className, info }: Props) {
  const { notifySuccess } = React.useContext(AppContext);
  const { path } = useRouteMatch();

  const sync = useSyncFluxObject({
    name: automation?.name,
    namespace: automation?.namespace,
    clusterName: automation?.clusterName,
    kind: automation?.kind,
  });

  const suspend = useToggleSuspend(
    {
      name: automation?.name,
      namespace: automation?.namespace,
      clusterName: automation?.clusterName,
      kind: automation?.kind,
      suspend: !automation?.suspended,
    },
    automation?.kind === FluxObjectKind.KindHelmRelease
      ? "helmrelease"
      : "kustomizations"
  );

  const handleSyncClicked = (opts) => {
    sync.mutateAsync(opts).then(() => {
      notifySuccess("Resource synced successfully");
    });
  };

  return (
    <Flex wide tall column className={className}>
      <DetailTitle name={automation?.name} type={automation?.kind} />
      {sync.isError && (
        <Alert
          severity="error"
          message={sync.error.message}
          title="Sync Error"
        />
      )}
      {suspend.isError && (
        <Alert
          severity="error"
          message={suspend.error.message}
          title="Sync Error"
        />
      )}
      <PageStatus
        conditions={automation?.conditions}
        suspended={automation?.suspended}
      />
<<<<<<< HEAD
      <Flex wide start>
        <SyncButton
          onClick={handleSyncClicked}
          loading={sync.isLoading}
          disabled={automation?.suspended}
        />
        <Spacer padding="xs" />
        <Button
          onClick={() => suspend.mutateAsync()}
          loading={suspend.isLoading}
        >
          {automation?.suspended ? "Resume" : "Suspend"}
        </Button>
      </Flex>

      <SubRouterTabs rootPath={`${path}/details`}>
        <RouterTab name="Details" path={`${path}/details`}>
          <>
            <InfoList items={info} />
            <ReconciledObjectsTable
=======
      <SyncButton
        onClick={handleSyncClicked}
        loading={sync.isLoading}
        disabled={automation?.suspended}
      />
      <p>
        <br></br>
        <br></br>
        <br></br>
        <br></br>
        <br></br>
        <br></br>
        <br></br>
        <br></br>
        <br></br>
        <br></br>
        <br></br>
        <br></br>
        <br></br>
        <br></br>
        <br></br>
        <br></br>
        <br></br>
        <br></br>
        <br></br>
        <br></br>
        <br></br>
        <br></br>
        <br></br>
        <br></br>
        <br></br>
      </p>
      <TabContent>
        <SubRouterTabs rootPath={`${path}/details`}>
          <RouterTab name="Details" path={`${path}/details`}>
            <>
              <InfoList items={info} />
              <ReconciledObjectsTable
                automationKind={automation?.kind}
                automationName={automation?.name}
                namespace={automation?.namespace}
                kinds={automation?.inventory}
                clusterName={automation?.clusterName}
              />
            </>
          </RouterTab>
          <RouterTab name="Events" path={`${path}/events`}>
            <EventsTable
              namespace={automation?.namespace}
              involvedObject={{
                kind: automation?.kind,
                name: automation?.name,
                namespace: automation?.namespace,
              }}
            />
          </RouterTab>
          <RouterTab name="Graph" path={`${path}/graph`}>
            <ReconciliationGraph
>>>>>>> 737ca32d (so close yet so far)
              automationKind={automation?.kind}
              automationName={automation?.name}
              namespace={automation?.namespace}
              kinds={automation?.inventory}
              clusterName={automation?.clusterName}
            />
          </>
        </RouterTab>
        <RouterTab name="Events" path={`${path}/events`}>
          <EventsTable
            namespace={automation?.namespace}
            involvedObject={{
              kind: automation?.kind,
              name: automation?.name,
              namespace: automation?.namespace,
            }}
          />
        </RouterTab>
        <RouterTab name="Graph" path={`${path}/graph`}>
          <ReconciliationGraph
            automationKind={automation?.kind}
            automationName={automation?.name}
            kinds={automation?.inventory}
            parentObject={automation}
            clusterName={automation?.clusterName}
            source={
              automation?.kind === FluxObjectKind.KindKustomization
                ? automation?.sourceRef
                : automation?.helmChart?.sourceRef
            }
          />
        </RouterTab>
      </SubRouterTabs>
    </Flex>
  );
}

export default styled(AutomationDetail).attrs({
  className: AutomationDetail.name,
})`
  ${PageStatus} {
    padding: ${(props) => props.theme.spacing.small} 0px;
  }
  ${SubRouterTabs} {
    margin-top: ${(props) => props.theme.spacing.medium};
  }
`;
