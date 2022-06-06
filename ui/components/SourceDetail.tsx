import _ from "lodash";
import * as React from "react";
import { useRouteMatch } from "react-router-dom";
import styled from "styled-components";
import { useListAutomations, useSyncFluxObject } from "../hooks/automations";
import { useToggleSuspend } from "../hooks/flux";
import { useListSources } from "../hooks/sources";
import { FluxObjectKind } from "../lib/api/core/types.pb";
import { AppContext } from "../contexts/AppContext";
import Alert from "./Alert";
import AutomationsTable from "./AutomationsTable";
import Button from "./Button";
import DetailTitle from "./DetailTitle";
import EventsTable from "./EventsTable";
import Flex from "./Flex";
import InfoList, { InfoField } from "./InfoList";
import LoadingPage from "./LoadingPage";
import PageStatus from "./PageStatus";
import Spacer from "./Spacer";
import SubRouterTabs, { RouterTab } from "./SubRouterTabs";
import SyncButton from "./SyncButton";

type Props = {
  className?: string;
  type: FluxObjectKind;
  name: string;
  namespace: string;
  clusterName: string;
  children?: JSX.Element;
  info: <T>(s: T) => InfoField[];
};

function SourceDetail({ className, name, namespace, clusterName, info, type }: Props) {
  const { notifySuccess } = React.useContext(AppContext);
  const { data: sources, isLoading, error } = useListSources();
  const { data: automations } = useListAutomations();
  const { path } = useRouteMatch();

  if (isLoading) {
    return <LoadingPage />;
  }

  const source = _.find(sources, { name, namespace, kind: type, clusterName });

  if (!source) {
    return (
      <Alert
        severity="error"
        title="Not found"
        message={`Could not find source '${name}'`}
      />
    );
  }

  const items = info(source);

  const isNameRelevant = (expectedName) => {
    return expectedName == name;
  };

  const isRelevant = (expectedType, expectedName) => {
    return expectedType == source.kind && isNameRelevant(expectedName);
  };

  const relevantAutomations = _.filter(automations, (a) => {
    if (!source) {
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
      name: source.name,
      namespace: source.namespace,
      clusterName: source.clusterName,
      kind: source.kind,
      suspend: !source.suspended,
    },
    "sources"
  );

  const sync = useSyncFluxObject({
    name: source?.name,
    namespace: source?.namespace,
    clusterName: source?.clusterName,
    kind: source?.kind,
  });

  const handleSyncClicked = () => {
    sync.mutateAsync({ withSource: false }).then(() => {
      notifySuccess("Resource synced successfully");
    });
  };

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
      <PageStatus conditions={source.conditions} suspended={source.suspended} />
      <Flex wide start>
        <SyncButton
          onClick={handleSyncClicked}
          loading={sync.isLoading}
          disabled={source?.suspended}
          hideDropdown={true}
        />
        <Spacer padding="xs" />
        <Button
          onClick={() => suspend.mutateAsync()}
          loading={suspend.isLoading}
        >
          {source.suspended ? "Resume" : "Suspend"}
        </Button>
      </Flex>

      <SubRouterTabs rootPath={`${path}/details`}>
        <RouterTab name="Details" path={`${path}/details`}>
          <>
            <InfoList items={items} />
            <AutomationsTable automations={relevantAutomations} hideSource />
          </>
        </RouterTab>
        <RouterTab name="Events" path={`${path}/events`}>
          <EventsTable
            namespace={source.namespace}
            involvedObject={{
              kind: type,
              name,
              namespace: source.namespace,
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
