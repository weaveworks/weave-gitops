import * as React from "react";
import styled from "styled-components";
import Interval from "../components/Interval";
import SourceDetail from "../components/SourceDetail";
import Timestamp from "../components/Timestamp";
import { Bucket, SourceRefSourceKind } from "../lib/api/core/types.pb";

type Props = {
  className?: string;
  name: string;
  namespace: string;
};

function BucketDetail({ name, namespace, className }: Props) {
  return (
    <SourceDetail
      className={className}
      name={name}
      namespace={namespace}
      type={SourceRefSourceKind.Bucket}
      // Guard against an undefined bucket with a default empty object
      info={(b: Bucket = {}) => [
        ["Endpoint", b.endpoint],
        ["Bucket Name", b.name],
        ["Last Updated", <Timestamp time={b.lastUpdatedAt} />],
        ["Interval", <Interval interval={b.interval} />],
        ["Cluster", b.clusterName],
        ["Namespace", b.namespace],
      ]}
    />
  );
}

export default styled(BucketDetail).attrs({ className: BucketDetail.name })``;
