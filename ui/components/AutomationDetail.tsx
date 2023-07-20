import * as React from "react";
import { useRouteMatch } from "react-router-dom";
import styled from "styled-components";
import { useSyncFluxObject } from "../hooks/automations";
import { useToggleSuspend } from "../hooks/flux";
import { createCanaryCondition, useGetInventory } from "../hooks/inventory";
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
import LargeInfo from "./LargeInfo";
import Metadata from "./Metadata";
import PageStatus from "./PageStatus";
import { PolicyViolationsList } from "./Policies/PolicyViolations/Table";
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
          name,
          namespace,
          clusterName,
          kind: Kind[type],
        },
      ],
      suspend: !automation.suspended,
    },
    "object"
  );

  const canaryStatus = createCanaryCondition(data?.objects);
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
              name,
              namespace,
              clusterName,
              kind: Kind[type],
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
    {
      name: "Violations",
      path: `${path}/violations`,
      component: () => {
        return (
          <PolicyViolationsList
            req={{
              application: name,
              clusterName,
              namespace,
              kind: type,
            }}
          />
        );
      },
      visible: true,
    },
    ...(customTabs?.length ? customTabs : []),
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
            <LargeInfo
              title={"Chart Version"}
              info={(automation as HelmRelease).helmChart?.version}
            />
          ) : (
            <LargeInfo
              title={"Applied Revision"}
              info={automation?.lastAppliedRevision}
            />
          )}
          <LargeInfo
            title={"Last Updated"}
            component={<Timestamp time={automationLastUpdated(automation)} />}
          />
        </Flex>
      </Flex>
      <PageStatus
        conditions={automation.conditions}
        suspended={automation.suspended}
      />
      {health && <HealthCheckAgg health={health} />}

      {(customTabs || customActions) && (
        <PageStatus conditions={[canaryStatus]} suspended={false} />
      )}

      <Collapsible>
        <div className="collapse-wrapper ">
          <div className="grid grid-items">
            {info.map(([k, v]) => {
              return (
                <Flex id={k} gap="8" key={k}>
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
        {defaultTabs
          .filter((r) => r.visible)
          .map((subRoute, index) => (
            <RouterTab name={subRoute.name} path={subRoute.path} key={index}>
              {subRoute.component()}
            </RouterTab>
          ))}
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
`;
