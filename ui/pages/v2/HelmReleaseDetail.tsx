import * as React from "react";
import styled from "styled-components";
import Heading from "../../components/Heading";
import InfoList from "../../components/InfoList";
import Interval from "../../components/Interval";
import KubeStatusIndicator from "../../components/KubeStatusIndicator";
import Link from "../../components/Link";
import Page from "../../components/Page";
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

  export type HelmRelease = {
    releaseName?: string;
    namespace?: string;
    name?: string;
    interval?: Interval;
    helmChart?: HelmChart;
    conditions?: Condition[];
    inventory?: GroupVersionKind[];
  };

  return (
    <Page loading={isLoading} error={error} className={className}>
      <Info>
        <Heading level={1}>{helmRelease?.name}</Heading>
        <Heading level={2}>{helmRelease?.namespace}</Heading>
        <InfoList
          items={[
            [
              "Source",
              <Link
                to={formatURL(V2Routes.GitRepo, {
                  name: helmRelease?.helmChart?.sourceRef.name,
                })}
              >
                {helmRelease?.helmChart?.sourceRef.kind}/
                {helmRelease?.helmChart?.sourceRef.name}
              </Link>,
            ],
            [
              "Status",
              <KubeStatusIndicator conditions={helmRelease?.conditions} />,
            ],
            ["Chart", helmRelease?.helmChart?.chart],
            ["Applied Revision", helmRelease?.helmChart?.version],
            ["Cluster", "Default"],
            ["Interval", <Interval interval={helmRelease?.interval} />],
          ]}
        />
      </Info>
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
