import * as React from "react";
import styled from "styled-components";
import { Button, Link } from "..";
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
};

function BucketDetail({ className, bucket }: Props) {
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

  const hasCreateRequestAnnotation =
    bucket.obj.metadata.annotations?.["templates.weave.works/create-request"];

  return (
    <SourceDetail
      className={className}
      type={Kind.Bucket}
      source={bucket}
      customActions={
        hasCreateRequestAnnotation && [
          <Link to={`/resources/${bucket.name}/edit`}>
            <Button id="edit-resource">Edit</Button>
          </Link>,
        ]
      }
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
