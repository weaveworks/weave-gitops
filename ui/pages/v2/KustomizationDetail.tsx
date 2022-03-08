import { Tab, Tabs } from "@material-ui/core";
import { createHashHistory } from "history";
import * as React from "react";
import { HashRouter, Redirect, Route, Switch } from "react-router-dom";
import styled from "styled-components";
import EventsTable from "../../components/EventsTable";
import Heading from "../../components/Heading";
import InfoList from "../../components/InfoList";
import Interval from "../../components/Interval";
import KubeStatusIndicator from "../../components/KubeStatusIndicator";
import Link from "../../components/Link";
import Page from "../../components/Page";
import ReconciledObjectsTable from "../../components/ReconciledObjectsTable";
import SourceLink from "../../components/SourceLink";
import Text from "../../components/Text";
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

function routesToIndex(route: string) {
  switch (route) {
    case "/details":
      return 0;
    case "/events":
      return 1;

    default:
      return 0;
  }
}

function activeClassName(route, hashHistory) {
  return `${hashHistory.location.pathname === route && "active-tab"}`;
}

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
      <Tabs
        indicatorColor="primary"
        value={routesToIndex(hashHistory.location.pathname)}
        onChange={(e, val) => {
          hashHistory.push(val === 0 ? "details" : "events");
        }}
      >
        <Link href="#/details">
          <Tab
            className={activeClassName("/details", hashHistory)}
            label={
              <Text uppercase bold>
                Details
              </Text>
            }
          />
        </Link>
        <Link href="#/events">
          <Tab
            className={activeClassName("/events", hashHistory)}
            label={
              <Text uppercase bold>
                Events
              </Text>
            }
          />
        </Link>
      </Tabs>
      <HashRouter>
        <Switch>
          <Route exact path="/details">
            <ReconciledObjectsTable
              kinds={kustomization?.inventory}
              automationName={kustomization?.name}
              namespace={WeGONamespace}
              automationKind={AutomationKind.KustomizationAutomation}
            />
          </Route>
          <Route exact path="/events">
            <EventsTable
              involvedObject={{
                kind: AutomationKind.KustomizationAutomation,
                name,
                namespace: kustomization?.namespace,
              }}
            />
          </Route>
          <Redirect exact from="/" to="/details" />
        </Switch>
      </HashRouter>
    </Page>
  );
}

export default styled(KustomizationDetail).attrs({
  className: KustomizationDetail.name,
})`
  .MuiTabs-root ${Link} .active-tab {
    background: ${(props) => props.theme.colors.primary}19;
  }
`;
