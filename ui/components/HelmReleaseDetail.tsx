import * as React from "react";
import styled from "styled-components";
import { useFeatureFlags } from "../hooks/featureflags";
import { Kind } from "../lib/api/core/types.pb";
import { HelmRelease } from "../lib/objects";
import Alert from "./Alert";
import AutomationDetail from "./AutomationDetail";
import ClusterDashboardLink from "./ClusterDashboardLink";
import { InfoField } from "./InfoList";
import Interval from "./Interval";
import { routeTab } from "./KustomizationDetail";
import SourceLink from "./SourceLink";

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
    />
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
