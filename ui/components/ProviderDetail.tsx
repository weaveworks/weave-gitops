import * as React from "react";
import { useRouteMatch } from "react-router-dom";
import styled from "styled-components";
import { useListAlerts } from "../hooks/notifications";
import { Provider } from "../lib/objects";
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
  const { path } = useRouteMatch();
  const { data, error } = useListAlerts(provider.provider, provider.namespace);
  return (
    <Flex column tall wide className={className}>
      <SubRouterTabs rootPath={`${path}/alerts`}>
        <RouterTab name="Alerts" path={`${path}/alerts`}>
          {error ? (
            <Alert severity="error" message={error.message} />
          ) : (
            <AlertsTable rows={data?.objects} />
          )}
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
