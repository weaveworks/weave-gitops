import * as React from "react";
import styled from "styled-components";
import { HelmRelease, SourceRefSourceKind } from "../lib/api/core/types.pb";
import { AutomationType } from "../lib/types";
import Alert from "./Alert";
import AutomationDetail from "./AutomationDetail";
import Interval from "./Interval";
import SourceLink from "./SourceLink";

type Props = {
  name: string;
  clusterName: string;
  helmRelease?: HelmRelease;
  className?: string;
};

function helmChartLink(helmRelease: HelmRelease) {
  if (helmRelease.helmChartName === "") {
    return (
      <SourceLink
        sourceRef={{
          kind: SourceRefSourceKind.HelmChart,
          name: helmRelease?.helmChart.chart,
        }}
      />
    );
  }

  const [ns, name] = helmRelease.helmChartName.split("/");

  return (
    <SourceLink
      sourceRef={{
        kind: SourceRefSourceKind.HelmChart,
        name: name,
        namespace: ns,
      }}
    />
  );
}

function HelmReleaseDetail({ helmRelease, className }: Props) {
  return (
    <AutomationDetail
      className={className}
      automation={{ ...helmRelease, type: AutomationType.HelmRelease }}
      info={[
        ["Source", helmChartLink(helmRelease)],
        ["Chart", helmRelease?.helmChart.chart],
        ["Cluster", helmRelease?.clusterName],
        ["Interval", <Interval interval={helmRelease?.interval} />],
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
