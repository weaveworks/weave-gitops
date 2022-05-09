import * as React from "react";
import styled from "styled-components";
import { Kustomization } from "../lib/api/core/types.pb";
import { AutomationType } from "../lib/types";
import Alert from "./Alert";
import AutomationDetail from "./AutomationDetail";
import Interval from "./Interval";
import SourceLink from "./SourceLink";

type Props = {
  kustomization?: Kustomization;
  className?: string;
};

function KustomizationDetail({ kustomization, className }: Props) {
  return (
    <AutomationDetail
      automation={{
        ...kustomization,
        type: AutomationType.Kustomization,
      }}
      info={[
        ["Source", <SourceLink sourceRef={kustomization?.sourceRef} />],
        ["Applied Revision", kustomization?.lastAppliedRevision],
        ["Cluster", kustomization?.clusterName],
        ["Path", kustomization?.path],
        ["Interval", <Interval interval={kustomization?.interval} />],
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
