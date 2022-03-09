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
import Link from "../../components/Link";
import Page from "../../components/Page";
import PageStatus from "../../components/PageStatus";
import ReconciledObjectsTable from "../../components/ReconciledObjectsTable";
import { useGetHelmRelease } from "../../hooks/automations";
import { AutomationKind } from "../../lib/api/core/types.pb";
import { formatURL } from "../../lib/nav";
import { V2Routes, WeGONamespace } from "../../lib/types";

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
                <Link
                  to={formatURL(V2Routes.HelmChart, {
                    name: helmRelease?.helmChart.name,
                  })}
                >
                  HelmChart/
                  {helmRelease?.helmChart.name}
                </Link>,
              ],
              ["Chart", helmRelease?.helmChart.chart],
              ["Cluster", "Default"],
              ["Interval", <Interval interval={helmRelease?.interval} />],
            ]}
          />
        </Info>
        <PageStatus ok={ok} msg={msg} error={error && true} />
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
