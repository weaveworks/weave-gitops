import * as React from "react";
import styled from "styled-components";
import { HelmRelease } from "../lib/objects";
import { FluxObjectKind } from "../lib/api/core/types.pb";
import { automationLastUpdated } from "../lib/utils";
import { useFeatureFlags } from "../hooks/featureflags";
import Alert from "./Alert";
import AutomationDetail from "./AutomationDetail";
import Interval from "./Interval";
import { routeTab } from "./KustomizationDetail";
import SourceLink from "./SourceLink";
import Timestamp from "./Timestamp";
import { InfoField } from "./InfoList";

type Props = {
  name: string;
  clusterName: string;
  helmRelease?: HelmRelease;
  className?: string;
  customTabs?: Array<routeTab>;
};

function helmChartLink(helmRelease: HelmRelease) {
  if (helmRelease?.helmChartName === "") {
    return (
      <SourceLink
        sourceRef={{
          kind: FluxObjectKind.KindHelmChart,
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
        kind: FluxObjectKind.KindHelmChart,
        name: name,
        namespace: ns,
      }}
      clusterName={helmRelease.clusterName}
    />
  );
}

function HelmReleaseDetail({ helmRelease, className, customTabs }: Props) {
  const { data } = useFeatureFlags();
  const flags = data?.flags || {};

  const tenancyInfo: InfoField[] =
    flags.WEAVE_GITOPS_FEATURE_TENANCY === "true" && helmRelease?.tenant
      ? [["Tenant", helmRelease?.tenant]]
      : [];
  const clusterInfo: InfoField[] =
    flags.WEAVE_GITOPS_FEATURE_CLUSTER === "true"
      ? [["Cluster", helmRelease?.clusterName]]
      : [];

  return (
    <AutomationDetail
      className={className}
      automation={helmRelease}
      customTabs={customTabs}
      info={[
        ["Source", helmChartLink(helmRelease)],
        ["Chart", helmRelease?.helmChart.chart],
        ...clusterInfo,
        ...tenancyInfo,
        ["Interval", <Interval interval={helmRelease?.interval} />],
        [
          "Last Updated",
          <Timestamp time={automationLastUpdated(helmRelease)} />,
        ],
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
