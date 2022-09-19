import * as React from "react";
import styled from "styled-components";
import Link from "../components/Link";
import SourceDetail from "../components/SourceDetail";
import Timestamp from "../components/Timestamp";
import { useFeatureFlags } from "../hooks/featureflags";
import { GitRepository, Kind } from "../lib/objects";
import { convertGitURLToGitProvider } from "../lib/utils";
import { InfoField } from "./InfoList";

type Props = {
  className?: string;
  gitRepository: GitRepository;
};

function GitRepositoryDetail({ className, gitRepository }: Props) {
  const { data } = useFeatureFlags();
  const flags = data?.flags || {};

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
      info={[
        ["Type", Kind.GitRepository],
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
