import * as React from "react";
import styled from "styled-components";
import { useListAlerts } from "../hooks/notifications";
import { Kind } from "../lib/api/core/types.pb";
import { Provider } from "../lib/objects";
import Alert from "./Alert";
import AlertsTable from "./AlertsTable";
import Flex from "./Flex";
import SubRouterTabs from "./SubRouterTabs";
import YamlView from "./YamlView";

type Props = {
  className?: string;
  provider?: Provider;
};

function ProviderDetail({ className, provider }: Props) {
  const { data, error } = useListAlerts(provider.provider, provider.namespace);
  const tabs = [
    {
      name: "Alerts",
      path: "alerts",
      component: () => {
        return error ? (
          <Alert severity="error" message={error.message} />
        ) : (
          <AlertsTable rows={data?.objects} />
        );
      },
    },
    {
      name: "Yaml",
      path: "yaml",
      component: () => {
        return (
          <YamlView
            object={{
              name: provider.name,
              namespace: provider.namespace,
              kind: Kind.Provider,
            }}
            yaml={provider.yaml}
          />
        );
      },
    },
  ];

  return (
    <Flex column tall wide className={className}>
      <SubRouterTabs tabs={tabs} />
    </Flex>
  );
}

export default styled(ProviderDetail).attrs({
  className: ProviderDetail.name,
})``;
