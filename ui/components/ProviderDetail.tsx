import * as React from "react";
import { useRouteMatch } from "react-router-dom";
import styled from "styled-components";
import { Kind, Provider } from "../lib/objects";
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
  console.log(path);
  console.log(provider);
  return (
    <Flex column tall wide className={className}>
      <SubRouterTabs rootPath={`${path}/alerts`}>
        <RouterTab name="Alerts" path={`${path}/alerts`}>
          <h1>in development</h1>
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
