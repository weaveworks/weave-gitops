import * as React from "react";
import styled from "styled-components";
import Interval from "../components/Interval";
import SourceDetail from "../components/SourceDetail";
import Timestamp from "../components/Timestamp";
import { removeKind } from "../lib/utils";
import { FluxObjectKind } from "../lib/api/core/types.pb";
import { HelmChart } from "../lib/objects";

type Props = {
  className?: string;
  name: string;
  namespace: string;
  clusterName: string;
};

function HelmChartDetail({ name, namespace, className, clusterName }: Props) {
  return (
    <SourceDetail
      name={name}
      namespace={namespace}
      type={FluxObjectKind.KindHelmChart}
      className={className}
      clusterName={clusterName}
      info={(ch: HelmChart = new HelmChart({})) => [
        ["Type", removeKind(FluxObjectKind.KindHelmChart)],
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

export default styled(HelmChartDetail).attrs({
  className: HelmChartDetail.name,
})``;
