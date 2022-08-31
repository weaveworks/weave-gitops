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
  helmRepository: HelmRepository;
};

function HelmRepositoryDetail({ className, helmRepository }: Props) {
  const { data } = useFeatureFlags();
  const flags = data?.flags || {};

  const tenancyInfo: InfoField[] =
    flags.WEAVE_GITOPS_FEATURE_TENANCY === "true" && helmRepository.tenant
      ? [["Tenant", helmRepository.tenant]]
      : [];
  const clusterInfo: InfoField[] =
    flags.WEAVE_GITOPS_FEATURE_CLUSTER === "true"
      ? [["Cluster", helmRepository.clusterName]]
      : [];

  return (
    <SourceDetail
      className={className}
      type={FluxObjectKind.KindHelmRepository}
      source={helmRepository}
      info={[
        ["Type", removeKind(FluxObjectKind.KindHelmRepository)],
        ["Repository Type", helmRepository.repositoryType.toLowerCase()],
        ["URL", <Link href={helmRepository.url}>{helmRepository.url}</Link>],
        [
          "Last Updated",
          helmRepository.lastUpdatedAt ? (
            <Timestamp time={helmRepository.lastUpdatedAt} />
          ) : (
            "-"
          ),
        ],
        ["Interval", <Interval interval={helmRepository.interval} />],
        ...clusterInfo,
        ["Namespace", helmRepository.namespace],
        ...tenancyInfo,
      ]}
    />
  );
}

export default styled(HelmRepositoryDetail).attrs({
  className: HelmRepositoryDetail.name,
})``;
