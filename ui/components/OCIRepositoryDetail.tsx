import * as React from "react";
import styled from "styled-components";
import { useFeatureFlags } from "../hooks/featureflags";
import { Kind } from "../lib/api/core/types.pb";
import { OCIRepository } from "../lib/objects";
import ClusterDashboardLink from "./ClusterDashboardLink";
import { InfoField } from "./InfoList";
import Interval from "./Interval";
import Link from "./Link";
import SourceDetail from "./SourceDetail";
import Timestamp from "./Timestamp";

type Props = {
  className?: string;
  ociRepository: OCIRepository;
  customActions?: JSX.Element[];
};

function OCIRepositoryDetail({
  className,
  ociRepository,
  customActions,
}: Props) {
  const { isFlagEnabled } = useFeatureFlags();

  const tenancyInfo: InfoField[] =
    isFlagEnabled("WEAVE_GITOPS_FEATURE_TENANCY") && ociRepository.tenant
      ? [["Tenant", ociRepository.tenant]]
      : [];
  const clusterInfo: InfoField[] = isFlagEnabled("WEAVE_GITOPS_FEATURE_CLUSTER")
    ? [
        [
          "Cluster",
          <ClusterDashboardLink clusterName={ociRepository.clusterName} />,
        ],
      ]
    : [];

  return (
    <SourceDetail
      className={className}
      type={Kind.OCIRepository}
      source={ociRepository}
      customActions={customActions}
      info={[
        ["Kind", Kind.OCIRepository],
        ["URL", <Link href={ociRepository.url}>{ociRepository.url}</Link>],
        [
          "Last Updated",
          ociRepository.lastUpdatedAt ? (
            <Timestamp time={ociRepository.lastUpdatedAt} />
          ) : (
            "-"
          ),
        ],
        ["Interval", <Interval interval={ociRepository.interval} />],
        ...clusterInfo,
        ["Namespace", ociRepository.namespace],
        [
          "Source",
          <Link href={ociRepository.source}>{ociRepository.source}</Link>,
        ],
        ["Revision", ociRepository.revision],
        ...tenancyInfo,
      ]}
    />
  );
}

export default styled(OCIRepositoryDetail).attrs({
  className: OCIRepositoryDetail.name,
})``;
