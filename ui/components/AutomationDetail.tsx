import { Dialog } from "@material-ui/core";
import * as React from "react";
import { useRouteMatch } from "react-router-dom";
import styled from "styled-components";
import { AppContext } from "../contexts/AppContext";
import { useSyncFluxObject } from "../hooks/automations";
import { useToggleSuspend } from "../hooks/flux";
import { Kind } from "../lib/api/core/types.pb";
import { Automation } from "../lib/objects";
import Button from "./Button";
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
  customActions?: any[];
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

  // default routes
  const defaultTabs: Array<routeTab> = [
    {
      name: "Details",
      path: `${path}/details`,
      component: () => {
        return (
          <>
            <InfoList items={info} />
            <Metadata metadata={automation.metadata} />
            <ReconciledObjectsTable automation={automation} />
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
            parentObject={automation}
            source={automation.sourceRef}
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
        {customActions?.map((action) => (
          <>
            <Spacer padding="xs" />
            {action}
          </>
        ))}
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
