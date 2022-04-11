import { createHashHistory } from "history";
import _ from "lodash";
import * as React from "react";
import styled from "styled-components";
import { useListAutomations } from "../hooks/automations";
import { useListSources } from "../hooks/sources";
import { SourceRefSourceKind } from "../lib/api/core/types.pb";
import Alert from "./Alert";
import AutomationsTable from "./AutomationsTable";
import EventsTable from "./EventsTable";
import Flex from "./Flex";
import HashRouterTabs, { HashRouterTab } from "./HashRouterTabs";
import Heading from "./Heading";
import InfoList, { InfoField } from "./InfoList";
import LoadingPage from "./LoadingPage";
import PageStatus from "./PageStatus";

type Props = {
  className?: string;
  type: SourceRefSourceKind;
  name: string;
  namespace: string;
  children?: JSX.Element;
  info: <T>(s: T) => InfoField[];
};

function SourceDetail({ className, name, info, type }: Props) {
  const { data: sources, isLoading, error } = useListSources();
  const { data: automations } = useListAutomations();

  if (isLoading) {
    return <LoadingPage />;
  }

  const s = _.find(sources, { name, type });

  if (!s) {
    return (
      <Alert
        severity="error"
        title="Not found"
        message={`Could not find source '${name}'`}
      />
    );
  }

  const items = info(s);

  const relevantAutomations = _.filter(automations, (a) => {
    if (!s) {
      return false;
    }

    if (a?.sourceRef?.kind == s.type && a?.sourceRef?.name == name) {
      return true;
    }

    return false;
  });

  return (
    <Flex wide tall column align className={className}>
      <Flex wide between>
        <div>
          <Heading level={2}>{s.type}</Heading>
          <InfoList items={items} />
        </div>
        <PageStatus conditions={s.conditions} suspended={s.suspended} />
      </Flex>
      {error && (
        <Alert severity="error" title="Error" message={error.message} />
      )}
      <HashRouterTabs history={createHashHistory()} defaultPath="/automations">
        <HashRouterTab name="Related Automations" path="/automations">
          <AutomationsTable automations={relevantAutomations} hideSource />
        </HashRouterTab>
        <HashRouterTab name="Events" path="/events">
          <EventsTable
            namespace={s.namespace}
            involvedObject={{
              kind: type,
              name,
              namespace: s.namespace,
            }}
          />
        </HashRouterTab>
      </HashRouterTabs>
    </Flex>
  );
}

export default styled(SourceDetail).attrs({ className: SourceDetail.name })`
  padding-top: ${(props) => props.theme.spacing.xs};
  ${InfoList} {
    margin-bottom: 60px;
  }
`;
