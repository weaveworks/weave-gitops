import * as React from "react";
import styled from "styled-components";
import Interval from "../components/Interval";
import SourceDetail from "../components/SourceDetail";
import Timestamp from "../components/Timestamp";
import { useFeatureFlags } from "../hooks/featureflags";
import { Kind } from "../lib/api/core/types.pb";
import { Bucket } from "../lib/objects";
import { InfoField } from "./InfoList";

type Props = {
  className?: string;
  bucket: Bucket;
  customActions?: JSX.Element[];
};

function BucketDetail({ className, bucket, customActions }: Props) {
  const { data } = useFeatureFlags();
  const flags = data?.flags || {};

  const tenancyInfo: InfoField[] =
    flags.WEAVE_GITOPS_FEATURE_TENANCY === "true" && bucket.tenant
      ? [["Tenant", bucket.tenant]]
      : [];

  const clusterInfo: InfoField[] =
    flags.WEAVE_GITOPS_FEATURE_CLUSTER === "true"
      ? [["Cluster", bucket.clusterName]]
      : [];

  return (
    <SourceDetail
      className={className}
      type={Kind.Bucket}
      source={bucket}
      customActions={customActions}
      info={[
        ["Type", Kind.Bucket],
        ["Endpoint", bucket.endpoint],
        ["Bucket Name", bucket.name],
        ["Last Updated", <Timestamp time={bucket.lastUpdatedAt} />],
        ["Interval", <Interval interval={bucket.interval} />],
        ...clusterInfo,
        ["Namespace", bucket.namespace],
        ...tenancyInfo,
      ]}
    />
  );
}

export default styled(BucketDetail).attrs({ className: BucketDetail.name })``;
