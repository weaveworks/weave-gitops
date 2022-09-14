import * as React from "react";
import { useRouteMatch } from "react-router-dom";
import styled from "styled-components";
import { Automation, useSyncFluxObject } from "../hooks/automations";
import { useToggleSuspend } from "../hooks/flux";
import { useGetObject } from "../hooks/objects";
import { FluxObjectKind } from "../lib/api/core/types.pb";
import { fluxObjectKindToKind } from "../lib/objects";
import { getSourceRefForAutomation } from "../lib/utils";
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
import YamlView from "./YamlView";

type Props = {
  automation?: Automation;
  className?: string;
  info: InfoField[];
  customTabs?: Array<routeTab>;
};

function AutomationDetail({ automation, className, info, customTabs }: Props) {
  const { path } = useRouteMatch();
  const { data: object } = useGetObject(
    automation?.name,
    automation?.namespace,
    fluxObjectKindToKind(automation?.kind),
    automation?.clusterName
  );

  const sync = useSyncFluxObject([
    {
      name: automation?.name,
      namespace: automation?.namespace,
      clusterName: automation?.clusterName,
      kind: automation?.kind,
    },
  ]);

  const suspend = useToggleSuspend(
    {
      objects: [
        {
          name: automation?.name,
          namespace: automation?.namespace,
          clusterName: automation?.clusterName,
          kind: automation?.kind,
        },
      ],
      suspend: !automation?.suspended,
    },
    automation?.kind === FluxObjectKind.KindHelmRelease
      ? "helmrelease"
      : "kustomizations"
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
            <Metadata metadata={object?.metadata} />
            <ReconciledObjectsTable
              automationKind={automation?.kind}
              automationName={automation?.name}
              namespace={automation?.namespace}
              kinds={automation?.inventory}
              clusterName={automation?.clusterName}
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
            namespace={automation?.namespace}
            involvedObject={{
              kind: automation?.kind,
              name: automation?.name,
              namespace: automation?.namespace,
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
            automationKind={automation?.kind}
            automationName={automation?.name}
            kinds={automation?.inventory}
            parentObject={automation}
            clusterName={automation?.clusterName}
            source={getSourceRefForAutomation(automation)}
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
            yaml={object.yaml}
            object={{
              kind: automation?.kind,
              name: automation?.name,
              namespace: automation?.namespace,
            }}
          />
        );
      },
      visible: !!object,
    },
  ];

  return (
    <Flex wide tall column className={className}>
      <Text size="large" semiBold titleHeight>
        {automation?.name}
      </Text>
      <PageStatus
        conditions={automation?.conditions}
        suspended={automation?.suspended}
      />
      <Flex wide start>
        <SyncButton
          onClick={(opts) => sync.mutateAsync(opts)}
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
