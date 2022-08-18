import * as React from "react";
import styled from "styled-components";
import Link from "../components/Link";
import SourceDetail from "../components/SourceDetail";
import Timestamp from "../components/Timestamp";
import { FluxObjectKind } from "../lib/api/core/types.pb";
import { convertGitURLToGitProvider, removeKind } from "../lib/utils";
import { GitRepository } from "../lib/objects";
import { useFeatureFlags } from "../hooks/featureflags";
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

  return (
    <SourceDetail
      className={className}
      type={FluxObjectKind.KindGitRepository}
      source={gitRepository}
      info={[
        ["Type", removeKind(FluxObjectKind.KindGitRepository)],
        [
          "URL",
          <Link newTab href={convertGitURLToGitProvider(gitRepository.url)}>
            {gitRepository.url}
          </Link>,
        ],
        ["Ref", gitRepository.reference?.branch],
        ["Last Updated", <Timestamp time={gitRepository.lastUpdatedAt} />],
        ["Cluster", gitRepository.clusterName],
        ["Namespace", gitRepository.namespace],
        ...tenancyInfo,
      ]}
    />
  );
}

export default styled(GitRepositoryDetail).attrs({
  className: GitRepositoryDetail.name,
})``;
