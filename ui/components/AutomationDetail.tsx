import { Dialog } from "@material-ui/core";
import * as React from "react";
import { useRouteMatch } from "react-router-dom";
import styled from "styled-components";
import { AppContext } from "../contexts/AppContext";
import { useSyncFluxObject } from "../hooks/automations";
import { useGetReconciledTree, useToggleSuspend } from "../hooks/flux";
import { Condition, Kind, ObjectRef } from "../lib/api/core/types.pb";
import { Automation, FluxObject } from "../lib/objects";
import { RequestError } from "../lib/types";
import Button from "./Button";
import CustomActions from "./CustomActions";
import DependenciesView from "./DependenciesView";
import EventsTable from "./EventsTable";
import Flex from "./Flex";
import InfoList, { InfoField } from "./InfoList";
import { routeTab } from "./KustomizationDetail";
import Metadata from "./Metadata";
import PageStatus from "./PageStatus";
import ReconciledObjectsTable from "./ReconciledObjectsTable";
import ReconciliationGraph from "./ReconciliationGraph";
import Spacer from "./Spacer";
import SubRouterTabs, { RouterTab } from "./SubRouterTabs";
import SyncButton from "./SyncButton";
import Text from "./Text";
import YamlView, { DialogYamlView } from "./YamlView";

type Props = {
  automation: Automation;
  className?: string;
  info: InfoField[];
  customTabs?: Array<routeTab>;
  customActions?: JSX.Element[];
};

export type ReconciledObjectsAutomation = {
  objects: FluxObject[] | undefined[];
  error?: RequestError;
  isLoading?: boolean;
  source: ObjectRef;
  name: string;
  namespace: string;
  suspended: boolean;
  conditions: Condition[];
  type: string;
  clusterName: string;
};

function AutomationDetail({
  automation,
  className,
  info,
  customTabs,
  customActions,
}: Props) {
  const { path } = useRouteMatch();
  const { setNodeYaml, appState } = React.useContext(AppContext);
  const nodeYaml = appState.nodeYaml;
  const sync = useSyncFluxObject([
    {
      name: automation.name,
      namespace: automation.namespace,
      clusterName: automation.clusterName,
      kind: Kind[automation.type],
    },
  ]);

  const suspend = useToggleSuspend(
    {
      objects: [
        {
          name: automation.name,
          namespace: automation.namespace,
          clusterName: automation.clusterName,
          kind: automation.type,
        },
      ],
      suspend: !automation.suspended,
    },
    automation.type === Kind.HelmRelease ? "helmrelease" : "kustomizations"
  );

  //grab data
  const {
    data: objects,
    error,
    isLoading,
  } = automation
    ? useGetReconciledTree(
        automation.name,
        automation.namespace,
        Kind[automation.type],
        automation.inventory,
        automation.clusterName
      )
    : { data: [], error: null, isLoading: false };
  const reconciledObjectsAutomation: ReconciledObjectsAutomation = {
    objects,
    error,
    isLoading,
    source: automation.sourceRef,
    name: automation.name,
    namespace: automation.namespace,
    suspended: automation.suspended,
    conditions: automation.conditions,
    type: automation.type,
    clusterName: automation.clusterName,
  };

  // default routes
  const defaultTabs: Array<routeTab> = [
    {
      name: "Details",
      path: `${path}/details`,
      component: () => {
        return (
          <>
            <InfoList items={info} />
            <Metadata
              metadata={automation.metadata}
              labels={automation.labels}
            />
            <ReconciledObjectsTable
              className={className}
              reconciledObjectsAutomation={reconciledObjectsAutomation}
            />
          </>
        );
      },
      visible: true,
    },
    {
      name: "Events",
      path: `${path}/events`,
      component: () => {
        return (
          <EventsTable
            namespace={automation.namespace}
            involvedObject={{
              kind: automation.type,
              name: automation.name,
              namespace: automation.namespace,
              clusterName: automation.clusterName,
            }}
          />
        );
      },
      visible: true,
    },
    {
      name: "Graph",
      path: `${path}/graph`,
      component: () => {
        return (
          <ReconciliationGraph
            className={className}
            reconciledObjectsAutomation={reconciledObjectsAutomation}
          />
        );
      },
      visible: true,
    },
    {
      name: "Dependencies",
      path: `${path}/dependencies`,
      component: () => <DependenciesView automation={automation} />,
      visible: true,
    },
    {
      name: "Yaml",
      path: `${path}/yaml`,
      component: () => {
        return (
          <YamlView
            yaml={automation.yaml}
            object={{
              kind: automation.type,
              name: automation.name,
              namespace: automation.namespace,
            }}
          />
        );
      },
      visible: true,
    },
  ];

  return (
    <Flex wide tall column className={className}>
      <Text size="large" semiBold titleHeight>
        {automation.name}
      </Text>
      <PageStatus
        conditions={automation.conditions}
        suspended={automation.suspended}
      />
      <Flex wide start>
        <SyncButton
          onClick={(opts) => sync.mutateAsync(opts)}
          loading={sync.isLoading}
          disabled={automation.suspended}
        />
        <Spacer padding="xs" />
        <Button
          onClick={() => suspend.mutateAsync()}
          loading={suspend.isLoading}
        >
          {automation.suspended ? "Resume" : "Suspend"}
        </Button>
        <CustomActions actions={customActions} />
      </Flex>

      <SubRouterTabs rootPath={`${path}/details`}>
        {defaultTabs.map(
          (subRoute, index) =>
            subRoute.visible && (
              <RouterTab name={subRoute.name} path={subRoute.path} key={index}>
                {subRoute.component()}
              </RouterTab>
            )
        )}
        {customTabs?.map(
          (customTab, index) =>
            customTab.visible && (
              <RouterTab
                name={customTab.name}
                path={customTab.path}
                key={index}
              >
                {customTab.component()}
              </RouterTab>
            )
        )}
      </SubRouterTabs>
      {nodeYaml && (
        <Dialog
          open={!!nodeYaml}
          onClose={() => setNodeYaml(null)}
          maxWidth="md"
          fullWidth
        >
          <DialogYamlView
            object={{
              name: nodeYaml.name,
              namespace: nodeYaml.namespace,
              kind: nodeYaml.type,
            }}
            yaml={nodeYaml.yaml}
          />
        </Dialog>
      )}
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
