import * as React from "react";
import styled from "styled-components";
import Flex from "../../components/Flex";
import Heading from "../../components/Heading";
import InfoList from "../../components/InfoList";
import Interval from "../../components/Interval";
import {
  computeMessage,
  computeReady,
} from "../../components/KubeStatusIndicator";
import Page from "../../components/Page";
import PageStatus from "../../components/PageStatus";
import ReconciledObjectsTable from "../../components/ReconciledObjectsTable";
import SourceLink from "../../components/SourceLink";
import { useGetHelmRelease } from "../../hooks/automations";
import {
  AutomationKind,
  SourceRefSourceKind,
} from "../../lib/api/core/types.pb";
import { WeGONamespace } from "../../lib/types";

type Props = {
  name: string;
  className?: string;
};

const Info = styled.div`
  padding-bottom: 32px;
`;

function HelmReleaseDetail({ className, name }: Props) {
  const { data, isLoading, error } = useGetHelmRelease(name);
  const helmRelease = data?.helmRelease;
  const ok = computeReady(helmRelease?.conditions);
  const msg = computeMessage(helmRelease?.conditions);

  return (
    <Page loading={isLoading} error={error} className={className}>
      <Flex wide between>
        <Info>
          <Heading level={1}>{helmRelease?.name}</Heading>
          <Heading level={2}>{helmRelease?.namespace}</Heading>
          <InfoList
            items={[
              [
                "Source",
                <SourceLink
                  sourceRef={{
                    kind: SourceRefSourceKind.HelmChart,
                    name: helmRelease.helmChart.chart,
                  }}
                />,
              ],
              ["Chart", helmRelease?.helmChart.chart],
              ["Cluster", "Default"],
              ["Interval", <Interval interval={helmRelease?.interval} />],
            ]}
          />
        </Info>
        <PageStatus ok={ok} msg={msg} error={error ? true : false} />
      </Flex>
      <ReconciledObjectsTable
        kinds={helmRelease?.inventory}
        automationName={helmRelease?.name}
        namespace={WeGONamespace}
        automationKind={AutomationKind.HelmReleaseAutomation}
      />
    </Page>
  );
}

export default styled(HelmReleaseDetail).attrs({
  className: HelmReleaseDetail.name,
})``;
