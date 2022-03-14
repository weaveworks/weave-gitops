import * as React from "react";
import styled from "styled-components";
import Page, { Content, TitleBar } from "../../components/Page";
import SourceDetail from "../../components/SourceDetail";
import {
  GitRepository,
  SourceRefSourceKind,
} from "../../lib/api/core/types.pb";

type Props = {
  className?: string;
  name: string;
  namespace: string;
};

function GitRepositoryDetail({ className, name, namespace }: Props) {
  return (
    <Page error={null} className={className}>
      <SourceDetail
        name={name}
        namespace={namespace}
        type={SourceRefSourceKind.GitRepository}
        info={(s: GitRepository) => [
          ["URL", s.url],
          ["Ref", s.reference.branch],
          ["Last Updated", ""],
          ["Cluster", "Default"],
          ["Namespace", s.namespace],
        ]}
      />
    </Page>
  );
}

export default styled(GitRepositoryDetail).attrs({
  className: GitRepositoryDetail.name,
})`
  ${TitleBar} {
    margin-bottom: 0;
  }

  ${Content} {
    padding-top: 0;
  }
`;
