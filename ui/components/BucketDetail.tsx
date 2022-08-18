import * as React from "react";
import styled from "styled-components";
import Interval from "../components/Interval";
import SourceDetail from "../components/SourceDetail";
import Timestamp from "../components/Timestamp";
import { FluxObjectKind } from "../lib/api/core/types.pb";
import { Bucket } from "../lib/objects";
import { removeKind } from "../lib/utils";
import { useFeatureFlags } from "../hooks/featureflags";
import { InfoField } from "./InfoList";

type Props = {
  className?: string;
  bucket: Bucket;
};

function BucketDetail({ className, bucket }: Props) {
  const { data } = useFeatureFlags();
  const flags = data?.flags || {};

  const tenancyInfo: InfoField[] =
    flags.WEAVE_GITOPS_FEATURE_TENANCY === "true" && bucket.tenant
      ? [["Tenant", bucket.tenant]]
      : [];

  return (
    <SourceDetail
      className={className}
      type={FluxObjectKind.KindBucket}
      source={bucket}
      info={[
        ["Type", removeKind(FluxObjectKind.KindBucket)],
        ["Endpoint", bucket.endpoint],
        ["Bucket Name", bucket.name],
        ["Last Updated", <Timestamp time={bucket.lastUpdatedAt} />],
        ["Interval", <Interval interval={bucket.interval} />],
        ["Cluster", bucket.clusterName],
        ["Namespace", bucket.namespace],
        ...tenancyInfo,
      ]}
    />
  );
}

export default styled(BucketDetail).attrs({ className: BucketDetail.name })``;
