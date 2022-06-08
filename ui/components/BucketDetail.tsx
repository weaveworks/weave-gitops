import * as React from "react";
import styled from "styled-components";
import Interval from "../components/Interval";
import SourceDetail from "../components/SourceDetail";
import Timestamp from "../components/Timestamp";
import { Bucket, FluxObjectKind } from "../lib/api/core/types.pb";
import { removeKind } from "../lib/utils";

type Props = {
  className?: string;
  name: string;
  namespace: string;
  clusterName: string;
};

function BucketDetail({ name, namespace, className, clusterName }: Props) {
  return (
    <SourceDetail
      className={className}
      name={name}
      namespace={namespace}
      clusterName={clusterName}
      type={FluxObjectKind.KindBucket}
      // Guard against an undefined bucket with a default empty object
      info={(b: Bucket = {}) => [
        ["Type", removeKind(FluxObjectKind.KindBucket)],
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
