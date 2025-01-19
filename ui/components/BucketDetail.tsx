import * as React from "react";
import type { JSX } from "react";
import styled from "styled-components";
import Interval from "../components/Interval";
import SourceDetail from "../components/SourceDetail";
import Timestamp from "../components/Timestamp";
import { useFeatureFlags } from "../hooks/featureflags";
import { Kind } from "../lib/api/core/types.pb";
import { Bucket } from "../lib/objects";
import ClusterDashboardLink from "./ClusterDashboardLink";
import { InfoField } from "./InfoList";

type Props = {
  className?: string;
  bucket: Bucket;
  customActions?: JSX.Element[];
};

function BucketDetail({ className, bucket, customActions }: Props) {
  const { isFlagEnabled } = useFeatureFlags();

  const tenancyInfo: InfoField[] =
    isFlagEnabled("WEAVE_GITOPS_FEATURE_TENANCY") && bucket.tenant
      ? [["Tenant", bucket.tenant]]
      : [];

  const clusterInfo: InfoField[] = isFlagEnabled("WEAVE_GITOPS_FEATURE_CLUSTER")
    ? [
        [
          "Cluster",
          <ClusterDashboardLink
            key={bucket.uid}
            clusterName={bucket?.clusterName}
          />,
        ],
      ]
    : [];

  return (
    <SourceDetail
      className={className}
      type={Kind.Bucket}
      source={bucket}
      customActions={customActions}
      info={[
        ["Kind", Kind.Bucket],
        ["Endpoint", bucket.endpoint],
        ["Bucket Name", bucket.name],
        [
          "Last Updated",
          <Timestamp key={bucket.uid} time={bucket.lastUpdatedAt} />,
        ],
        ["Interval", <Interval key={bucket.uid} interval={bucket.interval} />],
        ...clusterInfo,
        ["Namespace", bucket.namespace],
        ...tenancyInfo,
      ]}
    />
  );
}

export default styled(BucketDetail).attrs({ className: BucketDetail.name })``;
