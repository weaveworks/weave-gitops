import { Collapse } from "@material-ui/core";
import * as React from "react";
import { useRouteMatch } from "react-router-dom";
import styled from "styled-components";
import { useSyncFluxObject } from "../hooks/automations";
import { useToggleSuspend } from "../hooks/flux";
import { useGetInventory } from "../hooks/inventory";
import { Condition, Kind, ObjectRef } from "../lib/api/core/types.pb";
import { Automation, FluxObject } from "../lib/objects";
import Button from "./Button";
import CustomActions from "./CustomActions";
import DependenciesView from "./DependenciesView";
import EventsTable from "./EventsTable";
import Flex from "./Flex";
import { HealthStatusType } from "./HealthCheckStatusIndicator";
import Icon, { IconType } from "./Icon";
import InfoList, { InfoField } from "./InfoList";
import { routeTab } from "./KustomizationDetail";
import Metadata from "./Metadata";
import PageStatus from "./PageStatus";
import ReconciledObjectsTable from "./ReconciledObjectsTable";
import ReconciliationGraph from "./ReconciliationGraph";
import RequestStateHandler from "./RequestStateHandler";
import Spacer from "./Spacer";
import SubRouterTabs, { RouterTab } from "./SubRouterTabs";
import SyncButton from "./SyncButton";
import YamlView from "./YamlView";

type Props = {
  automation: Automation;
  className?: string;
  info: InfoField[];
  customTabs?: Array<routeTab>;
  customActions?: JSX.Element[];
  children?: JSX.Element;
};
const Collapsible = ({ children }) => {
  const [isOpen, setIsOpen] = React.useState(false);

  const toggle = () => setIsOpen(!isOpen);

  return (
    <div>
      <Collapse in={isOpen} collapsedSize={40}>
        {children}
      </Collapse>
      <div
        onClick={toggle}
        style={{
          display: "flex",
          justifyContent: "center",
          cursor: "pointer",
          background: "linear-gradient(to bottom, #ffffff, #f6f7f9)",
        }}
      >
        <Icon
          type={
            isOpen
              ? IconType.ArrowUpwardRoundedIcon
              : IconType.ArrowDownwardRoundedIcon
          }
          size="medium"
          text={isOpen ? "Show Less" : "Show More"}
        />
      </div>
    </div>
  );
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
interface AggStatus {
  health: {
    healthy: number;
    unhealthy: number;
    NA: number;
  };
}

function computeAggHealthCheck(objects: FluxObject[]): AggStatus {
  const healthAgg = { healthy: 0, unhealthy: 0, NA: 0 };
  objects.forEach(({ health }) => {
    switch (health.status) {
      case HealthStatusType.Healthy:
        healthAgg.healthy += 1;
        break;
      case HealthStatusType.Unhealthy:
        healthAgg.unhealthy += 1;
        break;
      default:
        healthAgg.NA += 1;
        break;
    }
  });
  return { health: healthAgg };
}

function AutomationDetail({
  automation,
  className,
  info,
  customTabs,
  customActions,
  children,
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
    false,
    {
      retry: false,
      refetchInterval: (data) =>
        data?.objects && location.href.includes("details") ? 5000 : false,
    }
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

  const { health } = computeAggHealthCheck(data?.objects || []);
  const defaultTabs: Array<routeTab> = [
    {
      name: "Details",
      path: `${path}/details`,
      component: () => {
        return (
          <>
            <Collapsible>
              <InfoList items={info} />
              <Metadata
                metadata={automation.metadata}
                labels={automation.labels}
              />
            </Collapsible>
            <RequestStateHandler loading={isLoading} error={error}>
              <ReconciledObjectsTable
                className={className}
                objects={data?.objects}
              />
            </RequestStateHandler>
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
      {children}
      {health && (
        <Flex wide gap="14">
          <span>Healthy:{health.healthy}</span>
          <span>UnHealthy:{health.unhealthy}</span>
          <span>NA:{health.NA}</span>
        </Flex>
      )}
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
