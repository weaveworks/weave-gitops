import * as React from "react";
import styled from "styled-components";
import Link from "../components/Link";
import SourceDetail from "../components/SourceDetail";
import Timestamp from "../components/Timestamp";
import { FluxObjectKind, GitRepository } from "../lib/api/core/types.pb";
import { convertGitURLToGitProvider } from "../lib/utils";

type Props = {
  className?: string;
  name: string;
  namespace: string;
};

function GitRepositoryDetail({ name, namespace, className }: Props) {
  return (
    <SourceDetail
      className={className}
      name={name}
      namespace={namespace}
      type={FluxObjectKind.KindGitRepository}
      info={(s: GitRepository) => [
        ["Type", FluxObjectKind.KindGitRepository],
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
