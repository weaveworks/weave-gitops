import * as React from "react";
import type { JSX } from "react";
import styled from "styled-components";
import Interval from "../components/Interval";
import SourceDetail from "../components/SourceDetail";
import Timestamp from "../components/Timestamp";
import { useFeatureFlags } from "../hooks/featureflags";
import { Kind } from "../lib/api/core/types.pb";
import { HelmChart } from "../lib/objects";
import ClusterDashboardLink from "./ClusterDashboardLink";
import { InfoField } from "./InfoList";

type Props = {
  className?: string;
  helmChart: HelmChart;
  customActions?: JSX.Element[];
};

function HelmChartDetail({ className, helmChart, customActions }: Props) {
  const { isFlagEnabled } = useFeatureFlags();

  const tenancyInfo: InfoField[] =
    isFlagEnabled("WEAVE_GITOPS_FEATURE_TENANCY") && helmChart.tenant
      ? [["Tenant", helmChart.tenant]]
      : [];
  const clusterInfo: InfoField[] = isFlagEnabled("WEAVE_GITOPS_FEATURE_CLUSTER")
    ? [
        [
          "Cluster",
          <ClusterDashboardLink
            key={helmChart.uid}
            clusterName={helmChart?.clusterName}
          />,
        ],
      ]
    : [];

  return (
    <SourceDetail
      type={Kind.HelmChart}
      className={className}
      source={helmChart}
      customActions={customActions}
      info={[
        ["Kind", Kind.HelmChart],
        ["Chart", helmChart.chart],
        ["Version", helmChart.version],
        ["Current Revision", helmChart.revision],
        ["Ref", helmChart.sourceRef?.name],
        [
          "Last Updated",
          <Timestamp key={helmChart.uid} time={helmChart.lastUpdatedAt} />,
        ],
        [
          "Interval",
          <Interval key={helmChart.uid} interval={helmChart.interval} />,
        ],
        ...clusterInfo,
        ["Namespace", helmChart.namespace],
        ...tenancyInfo,
      ]}
    />
  );
}

export default styled(HelmChartDetail).attrs({
  className: HelmChartDetail.name,
})``;
