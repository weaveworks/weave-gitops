import { createHashHistory } from "history";
import * as React from "react";
import styled from "styled-components";
import EventsTable from "../../components/EventsTable";
import HashRouterTabs, { HashRouterTab } from "../../components/HashRouterTabs";
import Heading from "../../components/Heading";
import InfoList from "../../components/InfoList";
import Interval from "../../components/Interval";
import KubeStatusIndicator from "../../components/KubeStatusIndicator";
import Page from "../../components/Page";
import ReconciledObjectsTable from "../../components/ReconciledObjectsTable";
import SourceLink from "../../components/SourceLink";
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
    <Page loading={isLoading} error={error} className={className}>
      <Info>
        <Heading level={1}>{kustomization?.name}</Heading>
        <Heading level={2}>{kustomization?.namespace}</Heading>
        <InfoList
          items={[
            ["Source", <SourceLink sourceRef={kustomization?.sourceRef} />],
            [
              "Status",
              <KubeStatusIndicator
                conditions={kustomization?.conditions}
                suspended={kustomization?.suspended}
              />,
            ],
            ["Applied Revision", kustomization?.lastAppliedRevision],
            ["Cluster", ""],
            ["Path", kustomization?.path],
            ["Interval", <Interval interval={kustomization?.interval} />],
            ["Last Updated At", kustomization?.lastHandledReconciledAt],
          ]}
        />
      </Info>
      <TabContent>
        <HashRouterTabs history={hashHistory} defaultPath="/details">
          <HashRouterTab name="Details" path="/details">
            <ReconciledObjectsTable
              kinds={kustomization?.inventory}
              automationName={kustomization?.name}
              namespace={WeGONamespace}
              automationKind={AutomationKind.KustomizationAutomation}
            />
          </HashRouterTab>
          <HashRouterTab name="Events" path="/events">
            <EventsTable
              involvedObject={{
                kind: AutomationKind.KustomizationAutomation,
                name,
                namespace: kustomization?.namespace,
              }}
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
