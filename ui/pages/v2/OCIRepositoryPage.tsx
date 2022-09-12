import * as React from "react";
import styled from "styled-components";
import OCIRepositoryDetail from "../../components/OCIRepositoryDetail";
import Page from "../../components/Page";
import { useGetObject } from "../../hooks/objects";
import { Kind, OCIRepository } from "../../lib/objects";

type Props = {
  className?: string;
  name: string;
  namespace: string;
  clusterName: string;
};

function OCIRepositoryPage({ className, name, namespace, clusterName }: Props) {
  const {
    data: ociRepository,
    isLoading,
    error,
  } = useGetObject<OCIRepository>(
    name,
    namespace,
    Kind.OCIRepository,
    clusterName
  );

  return (
    <Page error={error} loading={isLoading} className={className}>
      <OCIRepositoryDetail ociRepository={ociRepository} />
    </Page>
  );
}

export default styled(OCIRepositoryPage).attrs({
  className: OCIRepositoryPage.name,
})``;
