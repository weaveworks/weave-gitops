import { createHashHistory } from "history";
import * as React from "react";
import styled from "styled-components";
import EventsTable from "../../components/EventsTable";
import Flex from "../../components/Flex";
import HashRouterTabs, { HashRouterTab } from "../../components/HashRouterTabs";
import Heading from "../../components/Heading";
import InfoList from "../../components/InfoList";
import Interval from "../../components/Interval";
import Page from "../../components/Page";
import PageStatus from "../../components/PageStatus";
import ReconciledObjectsTable from "../../components/ReconciledObjectsTable";
import ReconciliationGraph from "../../components/ReconciliationGraph";
import SourceLink from "../../components/SourceLink";
import Spacer from "../../components/Spacer";
import { useGetKustomization } from "../../hooks/automations";
import { AutomationKind } from "../../lib/api/core/types.pb";
import { WeGONamespace } from "../../lib/types";

type Props = {
  name: string;
  className?: string;
};

const Info = styled.div`
  padding-bottom: 32px;
`;

const TabContent = styled.div`
  margin-top: 52px;
`;

function KustomizationDetail({ className, name }: Props) {
  const { data, isLoading, error } = useGetKustomization(name);
  const hashHistory = createHashHistory();
  const kustomization = data?.kustomization;
  return (
    <Page loading={isLoading} error={error} className={className} title={name}>
      <Spacer padding="xs" />
      <Flex wide between>
        <Info>
          <Heading level={2}>{kustomization?.namespace}</Heading>
          <InfoList
            items={[
              ["Source", <SourceLink sourceRef={kustomization?.sourceRef} />],
              ["Applied Revision", kustomization?.lastAppliedRevision],
              ["Cluster", "Default"],
              ["Path", kustomization?.path],
              ["Interval", <Interval interval={kustomization?.interval} />],
              ["Last Updated At", kustomization?.lastHandledReconciledAt],
            ]}
          />
        </Info>
        <PageStatus
          conditions={kustomization?.conditions}
          suspended={kustomization?.suspended}
        />
      </Flex>
      <TabContent>
        <HashRouterTabs history={hashHistory} defaultPath="/details">
          <HashRouterTab name="Details" path="/details">
            <ReconciledObjectsTable
              automationKind={AutomationKind.KustomizationAutomation}
              automationName={kustomization?.name}
              kinds={kustomization?.inventory}
              namespace={WeGONamespace}
            />
          </HashRouterTab>
          <HashRouterTab name="Events" path="/events">
            <EventsTable
              involvedObject={{
                kind: "Kustomization",
                name,
                namespace: kustomization?.namespace,
              }}
            />
          </HashRouterTab>
          <HashRouterTab name="Reconciliation Graph" path="/graph">
            <ReconciliationGraph
              automationKind={AutomationKind.KustomizationAutomation}
              automationName={kustomization?.name}
              kinds={kustomization?.inventory}
              parentObject={data?.kustomization}
              namespace={WeGONamespace}
            />
          </HashRouterTab>
        </HashRouterTabs>
      </TabContent>
    </Page>
  );
}

export default styled(KustomizationDetail).attrs({
  className: KustomizationDetail.name,
})``;
