import * as React from "react";
import styled from "styled-components";
import BucketDetailComponent from "../../components/BucketDetail";
import Page from "../../components/Page";

type Props = {
  className?: string;
  name: string;
  namespace: string;
  clusterName: string;
};

function BucketDetail({ className, name, namespace, clusterName }: Props) {
  return (
    <Page error={null} className={className}>
      <BucketDetailComponent name={name} namespace={namespace} clusterName={clusterName} />
    </Page>
  );
}

export default styled(BucketDetail).attrs({
  className: BucketDetail.name,
})``;
