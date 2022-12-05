import * as React from "react";
import styled from "styled-components";
import Link from "../components/Link";
import SourceDetail from "../components/SourceDetail";
import Timestamp from "../components/Timestamp";
import { useFeatureFlags } from "../hooks/featureflags";
import { Kind } from "../lib/api/core/types.pb";
import { GitRepository } from "../lib/objects";
import { convertGitURLToGitProvider } from "../lib/utils";
import { InfoField } from "./InfoList";

type Props = {
  className?: string;
  gitRepository: GitRepository;
  customActions?: JSX.Element[];
};

function GitRepositoryDetail({
  className,
  gitRepository,
  customActions,
}: Props) {
  const { data } = useFeatureFlags();
  const flags = data.flags;

  const tenancyInfo: InfoField[] =
    flags.WEAVE_GITOPS_FEATURE_TENANCY === "true" && gitRepository.tenant
      ? [["Tenant", gitRepository.tenant]]
      : [];
  const clusterInfo: InfoField[] =
    flags.WEAVE_GITOPS_FEATURE_CLUSTER === "true"
      ? [["Cluster", gitRepository.clusterName]]
      : [];

  return (
    <SourceDetail
      className={className}
      type={Kind.GitRepository}
      source={gitRepository}
      customActions={customActions}
      info={[
        ["Kind", Kind.GitRepository],
        [
          "URL",
          <Link newTab href={convertGitURLToGitProvider(gitRepository.url)}>
            {gitRepository.url}
          </Link>,
        ],
        ["Ref", gitRepository.reference?.branch],
        ["Last Updated", <Timestamp time={gitRepository.lastUpdatedAt} />],
        ...clusterInfo,
        ["Namespace", gitRepository.namespace],
        ...tenancyInfo,
      ]}
    />
  );
}

export default styled(GitRepositoryDetail).attrs({
  className: GitRepositoryDetail.name,
})``;
