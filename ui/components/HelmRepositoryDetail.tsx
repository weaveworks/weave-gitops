import * as React from "react";
import type { JSX } from "react";
import styled from "styled-components";
import { useFeatureFlags } from "../hooks/featureflags";
import { Kind } from "../lib/api/core/types.pb";
import { HelmRepository } from "../lib/objects";
import ClusterDashboardLink from "./ClusterDashboardLink";
import { InfoField } from "./InfoList";
import Interval from "./Interval";
import Link from "./Link";
import SourceDetail from "./SourceDetail";
import Timestamp from "./Timestamp";

type Props = {
  className?: string;
  helmRepository: HelmRepository;
  customActions?: JSX.Element[];
};

function HelmRepositoryDetail({
  className,
  helmRepository,
  customActions,
}: Props) {
  const { isFlagEnabled } = useFeatureFlags();

  const tenancyInfo: InfoField[] =
    isFlagEnabled("WEAVE_GITOPS_FEATURE_TENANCY") && helmRepository.tenant
      ? [["Tenant", helmRepository.tenant]]
      : [];
  const clusterInfo: InfoField[] = isFlagEnabled("WEAVE_GITOPS_FEATURE_CLUSTER")
    ? [
        [
          "Cluster",
          <ClusterDashboardLink
            key={helmRepository.uid}
            clusterName={helmRepository?.clusterName}
          />,
        ],
      ]
    : [];

  return (
    <SourceDetail
      className={className}
      type={Kind.HelmRepository}
      source={helmRepository}
      customActions={customActions}
      info={[
        ["Kind", Kind.HelmRepository],
        ["Repository Type", helmRepository.repositoryType.toLowerCase()],
        [
          "URL",
          <Link key={helmRepository.uid} href={helmRepository.url}>
            {helmRepository.url}
          </Link>,
        ],
        [
          "Last Updated",
          helmRepository.lastUpdatedAt ? (
            <Timestamp time={helmRepository.lastUpdatedAt} />
          ) : (
            "-"
          ),
        ],
        [
          "Interval",
          <Interval
            key={helmRepository.uid}
            interval={helmRepository.interval}
          />,
        ],
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
