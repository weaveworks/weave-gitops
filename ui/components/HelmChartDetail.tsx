import * as React from "react";
import styled from "styled-components";
import Interval from "../components/Interval";
import SourceDetail from "../components/SourceDetail";
import Timestamp from "../components/Timestamp";
import { useFeatureFlags } from "../hooks/featureflags";
import { Kind } from "../lib/api/core/types.pb";
import { HelmChart } from "../lib/objects";
import { InfoField } from "./InfoList";

type Props = {
  className?: string;
  helmChart: HelmChart;
  customActions?: JSX.Element[];
};

function HelmChartDetail({ className, helmChart, customActions }: Props) {
  const { data } = useFeatureFlags();
  const flags = data?.flags || {};

  const tenancyInfo: InfoField[] =
    flags.WEAVE_GITOPS_FEATURE_TENANCY === "true" && helmChart.tenant
      ? [["Tenant", helmChart.tenant]]
      : [];
  const clusterInfo: InfoField[] =
    flags.WEAVE_GITOPS_FEATURE_CLUSTER === "true"
      ? [["Cluster", helmChart.clusterName]]
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
        ["Last Updated", <Timestamp time={helmChart.lastUpdatedAt} />],
        ["Interval", <Interval interval={helmChart.interval} />],
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
