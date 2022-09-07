import * as React from "react";
import { useRouteMatch } from "react-router-dom";
import styled from "styled-components";
import { useFeatureFlags } from "../hooks/featureflags";
import { useListAlerts } from "../hooks/notifications";
import { CrossNamespaceObjectRef, Kind, Provider } from "../lib/objects";
import DataTable, { Field, filterConfig } from "./DataTable";
import EventsTable from "./EventsTable";
import Flex from "./Flex";
import SubRouterTabs, { RouterTab } from "./SubRouterTabs";
import YamlView from "./YamlView";

type Props = {
  className?: string;
  provider?: Provider;
};

function ProviderDetail({ className, provider }: Props) {
  const { path } = useRouteMatch();
  const { data, isLoading, error } = useListAlerts(
    provider.name,
    provider.namespace
  );
  const { data: flagData } = useFeatureFlags();
  const flags = flagData?.flags || {};

  let initialFilterState = {
    ...filterConfig(data, "name"),
    ...filterConfig(data, "namespace"),
    ...filterConfig(data, "eventSeverity"),
  };
  if (flags.WEAVE_GITOPS_FEATURE_CLUSTER === "true") {
    initialFilterState = {
      ...initialFilterState,
      ...filterConfig(data, "clusterName"),
    };
  }

  const alertFields: Field[] = [
    {
      label: "Name",
      value: "name",
    },
    {
      label: "Namespace",
      value: "namespace",
    },
    {
      label: "Severity",
      value: "eventSeverity",
    },
    {
      label: "Event Sources",
      value: (a) => {
        return (
          <ul>
            {a?.eventSources?.map((obj: CrossNamespaceObjectRef) => (
              <li>
                {obj.kind}: {obj.name}
              </li>
            ))}
          </ul>
        );
      },
      labelRenderer: (a) => <h2>Event Sources</h2>,
    },
    ...(flags.WEAVE_GITOPS_FEATURE_CLUSTER === "true"
      ? [{ label: "Cluster", value: (obj) => obj.clusterName }]
      : []),
  ];
  return (
    <Flex column tall wide className={className}>
      <SubRouterTabs rootPath={`${path}/alerts`}>
        <RouterTab name="Alerts" path={`${path}/alerts`}>
          <DataTable
            fields={alertFields}
            rows={data?.objects || []}
            filters={initialFilterState}
          />
        </RouterTab>
        <RouterTab name="Events" path={`${path}/events`}>
          <EventsTable
            involvedObject={{
              kind: Kind.Provider,
              name: provider?.name,
              namespace: provider?.namespace,
            }}
          />
        </RouterTab>
        <RouterTab name="Yaml" path={`${path}/yaml`}>
          <YamlView object={provider} yaml={provider.yaml} />
        </RouterTab>
      </SubRouterTabs>
    </Flex>
  );
}

export default styled(ProviderDetail).attrs({
  className: ProviderDetail.name,
})``;
