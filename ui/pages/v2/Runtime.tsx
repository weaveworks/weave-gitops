import * as React from "react";
import styled from "styled-components";
import FluxRuntimeComponent from "../../components/FluxRuntime";
import Page from "../../components/Page";
import { useListRuntimeCrds, useListRuntimeObjects } from "../../hooks/flux";

type Props = {
  className?: string;
};

function Runtime({ className }: Props) {
  const { data, isLoading, error } = useListRuntimeObjects();
  const {
    data: crds,
    isLoading: crdsLoading,
    error: crdsError,
  } = useListRuntimeCrds();
  return (
    <Page
      loading={isLoading || crdsLoading}
      error={error || crdsError}
      className={className}
      path={[{ label: "Runtime" }]}
    >
      <FluxRuntimeComponent deployments={data?.deployments} crds={crds?.crds} />
    </Page>
  );
}

export default styled(Runtime).attrs({ className: Runtime.name })``;
