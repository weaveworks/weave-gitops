import * as React from "react";
import styled from "styled-components";
import { useFeatureFlags } from "../../../hooks/featureflags";
import { useGetObject } from "../../../hooks/objects";
import { Kind } from "../../../lib/api/core/types.pb";
import { ImagePolicy } from "../../../lib/objects";
import { V2Routes } from "../../../lib/types";
import ClusterDashboardLink from "../../ClusterDashboardLink";
import Metadata from "../../Metadata";
import Page from "../../Page";
import ImageAutomationDetails from "../ImageAutomationDetails";

type Props = {
  className?: string;
  name: string;
  namespace: string;
  clusterName: string;
};

function ImagePolicyDetails({
  className,
  name,
  namespace,
  clusterName,
}: Props) {
  const { data, isLoading, error } = useGetObject<ImagePolicy>(
    name,
    namespace,
    Kind.ImagePolicy,
    clusterName,
    {
      refetchInterval: 5000,
    },
  );
  const { isFlagEnabled } = useFeatureFlags();
  const rootPath = V2Routes.ImagePolicyDetails;

  return (
    <Page
      error={error}
      loading={isLoading}
      className={className}
      path={[
        { label: "Image Policies", url: V2Routes.ImagePolicies },
        { label: name },
      ]}
    >
      {!!data && (
        <ImageAutomationDetails
          data={data}
          kind={Kind.ImagePolicy}
          infoFields={[
            {
              rowkey: "Kind",
              value: Kind.ImagePolicy,
            },
            {
              rowkey: "Namespace",
              value: data?.namespace,
            },
            {
              rowkey: "Cluster",
              children: <ClusterDashboardLink clusterName={clusterName} />,
              visible: isFlagEnabled("WEAVE_GITOPS_FEATURE_CLUSTER"),
            },
            {
              rowkey: "Image Policy",
              value: data?.imagePolicy?.type,
            },
            {
              rowkey: "Order/Range",
              value: data?.imagePolicy?.value,
            },
          ]}
          rootPath={rootPath}
        >
          <Metadata metadata={data?.metadata} labels={data?.labels} />
        </ImageAutomationDetails>
      )}
    </Page>
  );
}

export default styled(ImagePolicyDetails)``;
