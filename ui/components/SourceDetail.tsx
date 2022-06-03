import _ from "lodash";
import * as React from "react";
import { useRouteMatch } from "react-router-dom";
import styled from "styled-components";
import { useListAutomations } from "../hooks/automations";
import { useToggleSuspend } from "../hooks/flux";
import { useListSources } from "../hooks/sources";
import { FluxObjectKind } from "../lib/api/core/types.pb";
import Alert from "./Alert";
import AutomationsTable from "./AutomationsTable";
import Button from "./Button";
import DetailTitle from "./DetailTitle";
import EventsTable from "./EventsTable";
import Flex from "./Flex";
import InfoList, { InfoField } from "./InfoList";
import LoadingPage from "./LoadingPage";
import PageStatus from "./PageStatus";
import SubRouterTabs, { RouterTab } from "./SubRouterTabs";

type Props = {
  className?: string;
  type: FluxObjectKind;
  name: string;
  namespace: string;
  children?: JSX.Element;
  info: <T>(s: T) => InfoField[];
};

function SourceDetail({ className, name, namespace, info, type }: Props) {
  const { data: sources, isLoading, error } = useListSources();
  const { data: automations } = useListAutomations();
  const { path } = useRouteMatch();

  if (isLoading) {
    return <LoadingPage />;
  }

  const s = _.find(sources, { name, namespace, kind: type });

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

  const isNameRelevant = (expectedName) => {
    return expectedName == name;
  };

  const isRelevant = (expectedType, expectedName) => {
    return expectedType == s.kind && isNameRelevant(expectedName);
  };

  const relevantAutomations = _.filter(automations, (a) => {
    if (!s) {
      return false;
    }

    if (
      type == FluxObjectKind.KindHelmChart &&
      isNameRelevant(a?.helmChart?.name)
    ) {
      return true;
    }

    return (
      isRelevant(a?.sourceRef?.kind, a?.sourceRef?.name) ||
      isRelevant(a?.helmChart?.sourceRef?.kind, a?.helmChart?.sourceRef?.name)
    );
  });

  const suspend = useToggleSuspend(
    {
      name: s.name,
      namespace: s.namespace,
      clusterName: s.clusterName,
      kind: s.kind,
      suspend: !s.suspended,
    },
    "sources"
  );

  return (
    <Flex wide tall column className={className}>
      <DetailTitle name={name} type={type} />
      {error ||
        (suspend.error && (
          <Alert
            severity="error"
            title="Error"
            message={error.message || suspend.error.message}
          />
        ))}
      <PageStatus conditions={s.conditions} suspended={s.suspended} />
      <Button onClick={() => suspend.mutateAsync()} loading={suspend.isLoading}>
        {s.suspended ? "Resume" : "Suspend"}
      </Button>

      <SubRouterTabs rootPath={`${path}/details`}>
        <RouterTab name="Details" path={`${path}/details`}>
          <>
            <InfoList items={items} />
            <AutomationsTable automations={relevantAutomations} hideSource />
          </>
        </RouterTab>
        <RouterTab name="Events" path={`${path}/events`}>
          <EventsTable
            namespace={s.namespace}
            involvedObject={{
              kind: type,
              name,
              namespace: s.namespace,
            }}
          />
        </RouterTab>
      </SubRouterTabs>
    </Flex>
  );
}

export default styled(SourceDetail).attrs({ className: SourceDetail.name })`
  ${PageStatus} {
    padding: ${(props) => props.theme.spacing.small} 0px;
  }
  ${SubRouterTabs} {
    margin-top: ${(props) => props.theme.spacing.medium};
  }
`;
