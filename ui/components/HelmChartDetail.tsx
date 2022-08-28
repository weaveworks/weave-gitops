import * as React from "react";
import styled from "styled-components";
import Interval from "../components/Interval";
import SourceDetail from "../components/SourceDetail";
import Timestamp from "../components/Timestamp";
import { removeKind } from "../lib/utils";
import { FluxObjectKind } from "../lib/api/core/types.pb";
import { HelmChart } from "../lib/objects";
import { useFeatureFlags } from "../hooks/featureflags";
import { InfoField } from "./InfoList";

type Props = {
  className?: string;
  helmChart: HelmChart;
};

function HelmChartDetail({ className, helmChart }: Props) {
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
      type={FluxObjectKind.KindHelmChart}
      className={className}
      source={helmChart}
      info={[
        ["Type", removeKind(FluxObjectKind.KindHelmChart)],
        ["Chart", helmChart.chart],
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
