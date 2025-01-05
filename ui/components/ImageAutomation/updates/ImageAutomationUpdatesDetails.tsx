import * as React from "react";
import styled from "styled-components";
import { useFeatureFlags } from "../../../hooks/featureflags";
import { useGetObject } from "../../../hooks/objects";
import { Kind } from "../../../lib/api/core/types.pb";
import { ImageUpdateAutomation } from "../../../lib/objects";
import { V2Routes } from "../../../lib/types";
import ClusterDashboardLink from "../../ClusterDashboardLink";
import Metadata from "../../Metadata";
import Page from "../../Page";
import { RowItem } from "../../Policies/Utils/HeaderRows";
import SourceLink from "../../SourceLink";
import ImageAutomationDetails from "../ImageAutomationDetails";

type Props = {
  className?: string;
  name: string;
  namespace: string;
  clusterName: string;
};
function getInfoList(
  data: ImageUpdateAutomation,
  clusterName: string,
): RowItem[] {
  const {
    kind,
    spec: { update, git },
  } = data.obj;
  const { path } = update;
  const { commit, checkout, push } = git;
  const { isFlagEnabled } = useFeatureFlags();

  return [
    {
      rowkey: "Kind",
      value: kind,
    },
    {
      rowkey: "Namespace",
      value: data.namespace,
    },
    {
      rowkey: "Cluster",
      children: <ClusterDashboardLink clusterName={clusterName} />,
      visible: isFlagEnabled("WEAVE_GITOPS_FEATURE_CLUSTER"),
    },
    {
      rowkey: "Source",
      children: (
        <SourceLink sourceRef={data.sourceRef} clusterName={clusterName} />
      ),
    },
    {
      rowkey: "Update Path",
      value: path,
    },
    {
      rowkey: "Checkout Branch",
      value: checkout?.ref?.branch,
    },
    {
      rowkey: "Push Branch",
      value: push?.branch,
    },
    {
      rowkey: "Author Name",
      value: commit.author?.name,
    },
    {
      rowkey: "Author Email",
      value: commit.author?.email,
    },
    {
      rowkey: "Commit Template",
      children: <code> {commit.messageTemplate}</code>,
    },
  ];
}

function ImageAutomationUpdatesDetails({
  className,
  name,
  namespace,
  clusterName,
}: Props) {
  const { data, isLoading, error } = useGetObject<ImageUpdateAutomation>(
    name,
    namespace,
    Kind.ImageUpdateAutomation,
    clusterName,
    {
      refetchInterval: 5000,
    },
  );

  const rootPath = V2Routes.ImageAutomationUpdatesDetails;
  return (
    <Page
      error={error}
      loading={isLoading}
      className={className}
      path={[
        { label: "Image Updates", url: V2Routes.ImageUpdates },
        { label: name },
      ]}
    >
      {data && (
        <ImageAutomationDetails
          data={data}
          kind={Kind.ImageUpdateAutomation}
          infoFields={getInfoList(data, data.clusterName)}
          rootPath={rootPath}
        >
          <Metadata metadata={data.metadata} labels={data.labels} />
        </ImageAutomationDetails>
      )}
    </Page>
  );
}

export default styled(ImageAutomationUpdatesDetails)``;
