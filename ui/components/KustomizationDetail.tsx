import { createHashHistory } from "history";
import * as React from "react";
import styled from "styled-components";
import { Kustomization, AutomationKind } from "../lib/api/core/types.pb";
import { WeGONamespace } from "../lib/types";
import EventsTable from "./EventsTable";
import Flex from "./Flex";
import HashRouterTabs, { HashRouterTab } from "./HashRouterTabs";
import Heading from "./Heading";
import InfoList from "./InfoList";
import Interval from "./Interval";
import PageStatus from "./PageStatus";
import ReconciledObjectsTable from "./ReconciledObjectsTable";
import ReconciliationGraph from "./ReconciliationGraph";
import SourceLink from "./SourceLink";
import Spacer from "./Spacer";

type Props = {
  name: string,
  kustomization?: Kustomization,
}

const Info = styled.div`
  padding-bottom: 32px;
`;

const TabContent = styled.div`
  margin-top: 52px;
`;

export default function KustomizationDetail({ kustomization, name }: Props) {
  const hashHistory = createHashHistory();
  return (
    <>
      <Spacer padding="xs" />
      <Flex wide between>
        <Info>
          <Heading level={2}>{kustomization?.namespace}</Heading>
          <InfoList
            items={[
              ["Source", <SourceLink sourceRef={kustomization?.sourceRef} />],
              ["Applied Revision", kustomization?.lastAppliedRevision],
              ["Cluster", kustomization?.clusterName],
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
              clusterName={kustomization?.clusterName}
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
              parentObject={kustomization}
              namespace={WeGONamespace}
              clusterName={kustomization?.clusterName}
            />
          </HashRouterTab>
        </HashRouterTabs>
      </TabContent>
    </>
  );
}
