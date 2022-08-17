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
  name: string;
  namespace: string;
  clusterName: string;
};

function HelmChartDetail({ name, namespace, className, clusterName }: Props) {
  const { data } = useFeatureFlags();
  const flags = data?.flags || {};

  return (
    <SourceDetail
      name={name}
      namespace={namespace}
      type={FluxObjectKind.KindHelmChart}
      className={className}
      clusterName={clusterName}
      info={(ch: HelmChart = new HelmChart({})) =>
        [
          ["Type", removeKind(FluxObjectKind.KindHelmChart)],
          ["Chart", ch?.chart],
          ["Ref", ch?.sourceRef?.name],
          ["Last Updated", <Timestamp time={ch?.lastUpdatedAt} />],
          ["Interval", <Interval interval={ch?.interval} />],
          ["Cluster", ch?.clusterName],
          ["Namespace", ch?.namespace],
          ...(flags.WEAVE_GITOPS_FEATURE_TENANCY === "true" && ch.tenant
            ? [["Tenant", ch.tenant]]
            : []),
        ] as InfoField[]
      }
    />
  );
}

export default styled(HelmChartDetail).attrs({
  className: HelmChartDetail.name,
})``;
