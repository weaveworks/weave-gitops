import * as React from "react";
import styled from "styled-components";
import { Button } from "..";
import { useFeatureFlags } from "../hooks/featureflags";
import { Kind } from "../lib/api/core/types.pb";
import { HelmRepository } from "../lib/objects";
import EditButton from "./CustomActions";
import { InfoField } from "./InfoList";
import Interval from "./Interval";
import Link from "./Link";
import SourceDetail from "./SourceDetail";
import Timestamp from "./Timestamp";

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
      type={Kind.HelmRepository}
      source={helmRepository}
      customActions={[<EditButton resource={helmRepository} />]}
      info={[
        ["Type", Kind.HelmRepository],
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
