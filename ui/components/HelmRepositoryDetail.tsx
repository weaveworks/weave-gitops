import * as React from "react";
import styled from "styled-components";
import { removeKind } from "../lib/utils";
import { FluxObjectKind } from "../lib/api/core/types.pb";
import { HelmRepository } from "../lib/objects";
import { useFeatureFlags } from "../hooks/featureflags";
import Interval from "./Interval";
import Link from "./Link";
import SourceDetail from "./SourceDetail";
import Timestamp from "./Timestamp";
import { InfoField } from "./InfoList";

type Props = {
  className?: string;
  name: string;
  namespace: string;
  clusterName: string;
};

function HelmRepositoryDetail({
  name,
  namespace,
  className,
  clusterName,
}: Props) {
  const { data } = useFeatureFlags();
  const flags = data?.flags || {};

  return (
    <SourceDetail
      className={className}
      name={name}
      namespace={namespace}
      clusterName={clusterName}
      type={FluxObjectKind.KindHelmRepository}
      info={(hr: HelmRepository = new HelmRepository({})) =>
        [
          ["Type", removeKind(FluxObjectKind.KindHelmRepository)],
          ["Repository Type", hr.repositoryType.toLowerCase()],
          ["URL", <Link href={hr.url}>{hr.url}</Link>],
          [
            "Last Updated",
            hr.lastUpdatedAt ? <Timestamp time={hr.lastUpdatedAt} /> : "-",
          ],
          ["Interval", <Interval interval={hr.interval} />],
          ["Cluster", hr.clusterName],
          ["Namespace", hr.namespace],
          ...(flags.WEAVE_GITOPS_FEATURE_TENANCY === "true" && hr.tenant
            ? [["Tenant", hr.tenant]]
            : []),
        ] as InfoField[]
      }
    />
  );
}

export default styled(HelmRepositoryDetail).attrs({
  className: HelmRepositoryDetail.name,
})``;
