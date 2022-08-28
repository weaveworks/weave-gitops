import * as React from "react";
import styled from "styled-components";
import GitRepositoryDetailComponent from "../../components/GitRepositoryDetail";
import Page from "../../components/Page";
import { useGetObject } from "../../hooks/objects";
import { GitRepository, Kind } from "../../lib/objects";

type Props = {
  className?: string;
  name: string;
  namespace: string;
  clusterName: string;
};

function GitRepositoryDetail({
  className,
  name,
  namespace,
  clusterName,
}: Props) {
  const {
    data: gitRepository,
    isLoading,
    error,
  } = useGetObject<GitRepository>(
    name,
    namespace,
    Kind.GitRepository,
    clusterName
  );

  return (
    <Page error={error} loading={isLoading} className={className} title={name}>
      <GitRepositoryDetailComponent gitRepository={gitRepository} />
    </Page>
  );
}

export default styled(GitRepositoryDetail).attrs({
  className: GitRepositoryDetail.name,
})``;
