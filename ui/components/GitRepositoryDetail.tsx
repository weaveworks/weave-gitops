import * as React from "react";
import styled from "styled-components";
import Link from "../components/Link";
import SourceDetail from "../components/SourceDetail";
import Timestamp from "../components/Timestamp";
import { FluxObjectKind } from "../lib/api/core/types.pb";
import { convertGitURLToGitProvider, removeKind } from "../lib/utils";
import { GitRepository } from "../lib/objects";

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
  return (
    <SourceDetail
      className={className}
      name={name}
      namespace={namespace}
      clusterName={clusterName}
      type={FluxObjectKind.KindGitRepository}
      info={(s: GitRepository) => [
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
      ]}
    />
  );
}

export default styled(GitRepositoryDetail).attrs({
  className: GitRepositoryDetail.name,
})``;
