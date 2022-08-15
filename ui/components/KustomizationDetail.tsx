import * as React from "react";
import styled from "styled-components";
import { FluxObjectKind, Kustomization } from "../lib/api/core/types.pb";
import { automationLastUpdated } from "../lib/utils";
import { useFeatureFlags } from "../hooks/featureflags";
import Alert from "./Alert";
import AutomationDetail from "./AutomationDetail";
import Interval from "./Interval";
import SourceLink from "./SourceLink";
import Timestamp from "./Timestamp";

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
};

function KustomizationDetail({ kustomization, className, customTabs }: Props) {
  const { data } = useFeatureFlags();
  const flags = data?.flags || {};

  return (
    <AutomationDetail
      className={className}
      customTabs={customTabs}
      automation={{
        ...kustomization,
        kind: FluxObjectKind.KindKustomization,
      }}
      info={[
        [
          "Source",
          <SourceLink
            sourceRef={kustomization?.sourceRef}
            clusterName={kustomization?.clusterName}
          />,
        ],
        ["Applied Revision", kustomization?.lastAppliedRevision],
        ["Cluster", kustomization?.clusterName],
        flags.WEAVE_GITOPS_FEATURE_TENANCY === "true" &&
          kustomization?.tenant !== "" && ["Tenant", kustomization?.tenant],
        ["Path", kustomization?.path],
        ["Interval", <Interval interval={kustomization?.interval} />],
        [
          "Last Updated",
          <Timestamp time={automationLastUpdated(kustomization)} />,
        ],
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
