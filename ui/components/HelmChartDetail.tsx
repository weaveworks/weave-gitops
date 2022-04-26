import * as React from "react";
import styled from "styled-components";
import Interval from "../components/Interval";
import SourceDetail from "../components/SourceDetail";
import Timestamp from "../components/Timestamp";
import { HelmChart, SourceRefSourceKind } from "../lib/api/core/types.pb";

type Props = {
  className?: string;
  name: string;
  namespace: string;
};

function HelmChartDetail({ name, namespace, className }: Props) {
  return (
    <SourceDetail
      name={name}
      namespace={namespace}
      type={SourceRefSourceKind.HelmChart}
      className={className}
      info={(ch: HelmChart) => [
        ["Chart", ch?.chart],
        ["Ref", ch?.sourceRef?.name],
        ["Last Updated", <Timestamp time={ch?.lastUpdatedAt} />],
        ["Interval", <Interval interval={ch?.interval} />],
        ["Cluster", ch?.clusterName],
        ["Namespace", ch?.namespace],
      ]}
    />
  );
}

export default styled(HelmChartDetail).attrs({ className: HelmChartDetail.name })``;
