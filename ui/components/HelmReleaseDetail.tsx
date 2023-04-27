import * as React from "react";
import styled from "styled-components";
import { useFeatureFlags } from "../hooks/featureflags";
import { Kind } from "../lib/api/core/types.pb";
import { HelmRelease } from "../lib/objects";
import { automationLastUpdated } from "../lib/utils";
import Alert from "./Alert";
import AutomationDetail from "./AutomationDetail";
import ClusterDashboardLink from "./ClusterDashboardLink";
import Flex from "./Flex";
import { InfoField } from "./InfoList";
import Interval from "./Interval";
import { routeTab } from "./KustomizationDetail";
import SourceLink from "./SourceLink";
import Text from "./Text";
import Timestamp from "./Timestamp";

type Props = {
  name: string;
  clusterName: string;
  helmRelease?: HelmRelease;
  className?: string;
  customTabs?: Array<routeTab>;
  customActions?: JSX.Element[];
};

function helmChartLink(helmRelease: HelmRelease) {
  if (helmRelease?.helmChartName === "") {
    return (
      <SourceLink
        sourceRef={{
          kind: Kind.HelmChart,
          name: helmRelease?.helmChart.chart,
        }}
        clusterName={helmRelease?.clusterName}
      />
    );
  }

  const [ns, name] = helmRelease?.helmChartName
    ? helmRelease.helmChartName.split("/")
    : ["", ""];

  return (
    <SourceLink
      sourceRef={{
        kind: Kind.HelmChart,
        name: name,
        namespace: ns,
      }}
      clusterName={helmRelease.clusterName}
    />
  );
}

function HelmReleaseDetail({
  helmRelease,
  className,
  customTabs,
  customActions,
}: Props) {
  const { isFlagEnabled } = useFeatureFlags();

  const tenancyInfo: InfoField[] =
    isFlagEnabled("WEAVE_GITOPS_FEATURE_TENANCY") && helmRelease?.tenant
      ? [["Tenant", helmRelease?.tenant]]
      : [];
  const clusterInfo: InfoField[] = isFlagEnabled("WEAVE_GITOPS_FEATURE_CLUSTER")
    ? [
        [
          "Cluster",
          <ClusterDashboardLink clusterName={helmRelease?.clusterName} />,
        ],
      ]
    : [];

  return (
    <AutomationDetail
      className={className}
      automation={helmRelease}
      customTabs={customTabs}
      customActions={customActions}
      info={[
        ["Kind", Kind.HelmRelease],
        ["Source", helmChartLink(helmRelease)],
        ["Chart", helmRelease?.helmChart.chart],
        ["Last Applied Revision", helmRelease.lastAppliedRevision],
        ["Last Attempted Revision", helmRelease.lastAttemptedRevision],
        ...clusterInfo,
        ...tenancyInfo,
        ["Interval", <Interval interval={helmRelease?.interval} />],
        ["Namespace", helmRelease?.namespace],
      ]}
    >
      <Flex wide end gap="14">
        <Text capitalize semiBold color="neutral30">
          Chart Version:{" "}
          <Text size="large" color="neutral40">{helmRelease.helmChart.version}</Text>
        </Text>
        <Text capitalize semiBold color="neutral30">
          Last Updated:{" "}
          <Text size="large" color="neutral40">
            <Timestamp time={automationLastUpdated(helmRelease)} />
          </Text>
        </Text>
      </Flex>
    </AutomationDetail>
  );
}

export default styled(HelmReleaseDetail).attrs({
  className: HelmReleaseDetail.name,
})`
  width: 100%;

  ${Alert} {
    margin-bottom: 16px;
  }
`;
