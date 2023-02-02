import * as React from "react";
import styled from "styled-components";
import { useGetObject } from "../../../hooks/objects";
import { Kind } from "../../../lib/api/core/types.pb";
import { ImageUpdateAutomation } from "../../../lib/objects";
import { V2Routes } from "../../../lib/types";
import { InfoField } from "../../InfoList";
import Metadata from "../../Metadata";
import Page from "../../Page";
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
  clusterName: string
): InfoField[] {
  const {
    kind,
    spec: { update, git },
  } = data.obj;
  const { path } = update;
  const { commit, checkout, push } = git;

  return [
    ["Kind", kind],
    ["Namespace", data.namespace],
    [
      "Source",
      <SourceLink sourceRef={data.sourceRef} clusterName={clusterName} />,
    ],
    ["Update Path", path],
    ["Checkout Branch", checkout?.ref?.branch],
    ["Push Branch", push?.branch],
    ["Author Name", commit.author?.name],
    ["Author Email", commit.author?.email],
    ["Commit Template", <code> {commit.messageTemplate}</code>],
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
    }
  );

  const rootPath = V2Routes.ImageAutomationUpdatesDetails;
  return (
    <Page error={error} loading={isLoading} className={className}>
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
