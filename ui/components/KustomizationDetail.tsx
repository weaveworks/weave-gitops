import * as React from "react";
import type { JSX } from "react";
import styled from "styled-components";
import { useFeatureFlags } from "../hooks/featureflags";
import { Kind } from "../lib/api/core/types.pb";
import { Kustomization } from "../lib/objects";
import Alert from "./Alert";
import AutomationDetail from "./AutomationDetail";
import ClusterDashboardLink from "./ClusterDashboardLink";
import { InfoField } from "./InfoList";
import Interval from "./Interval";
import SourceLink from "./SourceLink";

export interface routeTab {
  name: string;
  path: string;
  visible?: boolean;
  component: (param?: any) => any;
}

type Props = {
  kustomization?: Kustomization;
  className?: string;
  customTabs?: Array<routeTab>;
  customActions?: JSX.Element[];
};

function KustomizationDetail({
  kustomization,
  className,
  customTabs,
  customActions,
}: Props) {
  const { isFlagEnabled } = useFeatureFlags();

  const tenancyInfo: InfoField[] =
    isFlagEnabled("WEAVE_GITOPS_FEATURE_TENANCY") && kustomization?.tenant
      ? [["Tenant", kustomization?.tenant]]
      : [];

  const clusterInfo: InfoField[] = isFlagEnabled("WEAVE_GITOPS_FEATURE_CLUSTER")
    ? [
        [
          "Cluster",
          <ClusterDashboardLink
            key={kustomization.uid}
            clusterName={kustomization?.clusterName}
          />,
        ],
      ]
    : [];

  return (
    <AutomationDetail
      className={className}
      customTabs={customTabs}
      automation={kustomization}
      customActions={customActions}
      info={[
        ["Kind", Kind.Kustomization],
        [
          "Source",
          <SourceLink
            key={kustomization.uid}
            sourceRef={kustomization?.sourceRef}
            clusterName={kustomization?.clusterName}
          />,
        ],
        ...clusterInfo,
        ...tenancyInfo,
        ["Path", kustomization?.path],
        [
          "Interval",
          <Interval
            key={kustomization.uid}
            interval={kustomization?.interval}
          />,
        ],
        ["Namespace", kustomization?.namespace],
      ]}
    />
  );
}

export default styled(KustomizationDetail).attrs({
  className: KustomizationDetail.name,
})`
  width: 100%;

  ${Alert} {
    margin-bottom: 16px;
  }
`;
