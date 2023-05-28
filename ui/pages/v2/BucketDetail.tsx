import * as React from "react";
import styled from "styled-components";
import BucketDetailComponent from "../../components/BucketDetail";
import Page from "../../components/Page";
import { useGetObject } from "../../hooks/objects";
import { Kind } from "../../lib/api/core/types.pb";
import { Bucket } from "../../lib/objects";
import { V2Routes } from "../../lib/types";

type Props = {
  className?: string;
  name: string;
  namespace: string;
  clusterName: string;
};

function BucketDetail({ className, name, namespace, clusterName }: Props) {
  const {
    data: bucket,
    isLoading,
    error,
  } = useGetObject<Bucket>(name, namespace, Kind.Bucket, clusterName);

  return (
    <Page
      error={error}
      loading={isLoading}
      className={className}
      path={[{ label: "Sources", url: V2Routes.Sources }, { label: name }]}
    >
      <BucketDetailComponent bucket={bucket} />
    </Page>
  );
}

export default styled(BucketDetail).attrs({
  className: BucketDetail.name,
})``;
