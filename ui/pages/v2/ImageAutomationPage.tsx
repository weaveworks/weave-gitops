import * as React from "react";
import styled from "styled-components";
import Page from "../../components/Page";
import { useListImageAutomation } from "../../hooks/imageautomation";
import { Kind } from "../../lib/api/core/types.pb";

type Props = {
  className?: string;
};

function ImageAutomation({ className }: Props) {
  const { data, isLoading, error } = useListImageAutomation(Kind.ImageAutomation);
 
  return (
    <Page
      loading={isLoading }
      error={error }
      className={className}
    >
      {/* <FluxRuntimeComponent deployments={data?.deployments} crds={crds?.crds} /> */}
    </Page>
  );
}

export default styled(ImageAutomation).attrs({className: ImageAutomation.name})``;
