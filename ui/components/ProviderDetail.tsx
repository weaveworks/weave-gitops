import * as React from "react";
import styled from "styled-components";
import { useListAlerts } from "../hooks/notifications";
import { Kind } from "../lib/api/core/types.pb";
import { Provider } from "../lib/objects";
import { createYamlCommand } from "../lib/utils";
import Alert from "./Alert";
import AlertsTable from "./AlertsTable";
import Flex from "./Flex";
import SubRouterTabs, { RouterTab } from "./SubRouterTabs";
import YamlView from "./YamlView";

type Props = {
  className?: string;
  provider?: Provider;
};

function ProviderDetail({ className, provider }: Props) {
  const { data, error } = useListAlerts(provider.name, provider.namespace);
  return (
    <Flex column tall wide className={className}>
      <SubRouterTabs rootPath="alerts">
        <RouterTab name="Alerts" path="alerts">
          {error ? (
            <Alert severity="error" message={error.message} />
          ) : (
            <AlertsTable rows={data?.objects} />
          )}
        </RouterTab>
        <RouterTab name="Yaml" path="yaml">
          <YamlView
            header={createYamlCommand(
              Kind.Provider,
              provider.name,
              provider.namespace,
            )}
            yaml={provider.yaml}
          />
        </RouterTab>
      </SubRouterTabs>
    </Flex>
  );
}

export default styled(ProviderDetail).attrs({
  className: ProviderDetail.name,
})``;
