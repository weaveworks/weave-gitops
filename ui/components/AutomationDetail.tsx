import { Tooltip } from "@material-ui/core";
import * as React from "react";
import { useRouteMatch } from "react-router-dom";
import styled from "styled-components";
import { useSyncFluxObject } from "../hooks/automations";
import { useToggleSuspend } from "../hooks/flux";
import { useGetInventory } from "../hooks/inventory";
import { Condition, Kind, ObjectRef } from "../lib/api/core/types.pb";
import { Automation, HelmRelease } from "../lib/objects";
import { automationLastUpdated } from "../lib/utils";
import Button from "./Button";
import Collapsible from "./Collapsible";
import CustomActions from "./CustomActions";
import DependenciesView from "./DependenciesView";
import EventsTable from "./EventsTable";
import Flex from "./Flex";
import HealthCheckAgg, { computeAggHealthCheck } from "./HealthCheckAgg";
import { InfoField } from "./InfoList";
import { routeTab } from "./KustomizationDetail";
import Metadata from "./Metadata";
import PageStatus from "./PageStatus";
import ReconciledObjectsTable from "./ReconciledObjectsTable";
import ReconciliationGraph from "./ReconciliationGraph";
import RequestStateHandler from "./RequestStateHandler";
import Spacer from "./Spacer";
import SubRouterTabs, { RouterTab } from "./SubRouterTabs";
import SyncButton from "./SyncButton";
import Text from "./Text";
import Timestamp from "./Timestamp";
import YamlView from "./YamlView";

type Props = {
  automation: Automation;
  className?: string;
  info: InfoField[];
  customTabs?: Array<routeTab>;
  customActions?: JSX.Element[];
};

export type ReconciledObjectsAutomation = {
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
  const {
    name,
    namespace,
    clusterName,
    type,
    suspended,
    conditions,
    sourceRef,
  } = automation;
  const reconciledObjectsAutomation: ReconciledObjectsAutomation = {
    name,
    namespace,
    clusterName,
    type: Kind[type],
    suspended,
    conditions,
    source: sourceRef,
  };
  const { path } = useRouteMatch();
  const { data, isLoading, error } = useGetInventory(
    type,
    name,
    clusterName,
    namespace,
    false
  );

  const sync = useSyncFluxObject([
    {
      name,
      namespace,
      clusterName,
      kind: Kind[type],
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

  // agreed to hide canary status agg for now as there's some concerns about the pd status ( when it's ready and when it's not)
  // const canaryStatus = createCanaryCondition(data?.objects);
  const health = computeAggHealthCheck(data?.objects || []);

  const defaultTabs: Array<routeTab> = [
    {
      name: "Details",
      path: `${path}/details`,
      component: () => {
        return (
          <RequestStateHandler loading={isLoading} error={error}>
            <ReconciledObjectsTable
              className={className}
              objects={data?.objects}
            />
          </RequestStateHandler>
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
    <Flex wide tall column className={className} gap="16">
      <Flex wide>
        <Flex start>
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
        <Flex wide end gap="14">
          {automation?.type === "HelmRelease" ? (
            <Text capitalize semiBold color="neutral30">
              Chart Version:{" "}
              <Text size="large" color="neutral40">
                {(automation as HelmRelease).helmChart?.version || "-"}
              </Text>
            </Text>
          ) : (
            <Flex gap="4" alignItems="baseline">
              <Text capitalize semiBold color="neutral30">
                Applied Revision:
              </Text>
              <Tooltip
                title={automation?.lastAppliedRevision || "-"}
                placement="top"
              >
                <Text size="large" color="neutral40" className="trim-text">
                  {automation?.lastAppliedRevision || "-"}
                </Text>
              </Tooltip>
            </Flex>
          )}
          <Text capitalize semiBold color="neutral30">
            Last Updated:{" "}
            <Text size="large" color="neutral40">
              <Timestamp time={automationLastUpdated(automation)} />
            </Text>
          </Text>
        </Flex>
      </Flex>
      <PageStatus
        conditions={automation.conditions}
        suspended={automation.suspended}
      />
      {health && <HealthCheckAgg health={health} />}

      {/* {(customTabs || customActions) && (
        <PageStatus conditions={[canaryStatus]} suspended={false} />
      )} */}

      <Collapsible>
        <div className="collapse-wrapper ">
          <div className="grid grid-items">
            {info.map(([k, v]) => {
              return (
                <Flex id={k} gap="8">
                  <Text capitalize semiBold color="neutral30">
                    {k}:
                  </Text>
                  {v || "-"}
                </Flex>
              );
            })}
          </div>
          <Metadata metadata={automation.metadata} labels={automation.labels} />
        </div>
      </Collapsible>

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
  ${Collapsible} {
    width: 100%;
  }
  ${PageStatus} {
    padding: ${(props) => props.theme.spacing.small} 0px;
  }
  .collapse-wrapper {
    padding: 16px 44px;
    width: 100%;
  }
  .grid {
    width: 100%;
    display: grid;
    gap: 8px;
  }
  .grid-items {
    grid-template-columns: repeat(auto-fit, minmax(calc(50% - 8px), 1fr));
  }
  .trim-text {
    max-width: 150px;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }
`;
