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
  name: string;
  namespace: string;
  clusterName: string;
};

function GitRepositoryDetail({
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
      type={FluxObjectKind.KindGitRepository}
      info={(s: GitRepository) =>
        [
          ["Type", removeKind(FluxObjectKind.KindGitRepository)],
          [
            "URL",
            <Link newTab href={convertGitURLToGitProvider(s.url)}>
              {s.url}
            </Link>,
          ],
          ["Ref", s.reference.branch],
          ["Last Updated", <Timestamp time={s.lastUpdatedAt} />],
          ["Cluster", s.clusterName],
          ["Namespace", s.namespace],
          ...(flags.WEAVE_GITOPS_FEATURE_TENANCY === "true" && s.tenant
            ? [["Tenant", s.tenant]]
            : []),
        ] as InfoField[]
      }
    />
  );
}

export default styled(GitRepositoryDetail).attrs({
  className: GitRepositoryDetail.name,
})``;
