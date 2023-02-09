import * as React from "react";
import styled from "styled-components";
import { useGetObject } from "../../../hooks/objects";
import { Kind } from "../../../lib/api/core/types.pb";
import { ImagePolicy } from "../../../lib/objects";
import { V2Routes } from "../../../lib/types";
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
    }
  );
  const rootPath = V2Routes.ImagePolicyDetails;
  return (
    <Page error={error} loading={isLoading} className={className}>
      {!!data && (
        <ImageAutomationDetails
          data={data}
          kind={Kind.ImagePolicy}
          infoFields={[
            ["Kind", Kind.ImagePolicy],
            ["Namespace", data?.namespace],
            ["Image Policy", data?.imagePolicy?.type || ""],
            ["Order/Range", data?.imagePolicy?.value],
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
