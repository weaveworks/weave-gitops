import * as React from "react";
import styled from "styled-components";
import FluxRuntimeComponent from "../../components/FluxRuntime";
import Page from "../../components/Page";
import { useListFluxCrds, useListFluxRuntimeObjects } from "../../hooks/flux";

type Props = {
  className?: string;
};

function FluxRuntime({ className }: Props) {
  const { data, isLoading, error } = useListFluxRuntimeObjects();
  const {
    data: crds,
    isLoading: crdsLoading,
    error: crdsError,
  } = useListFluxCrds();
  return (
    <Page
      loading={isLoading || crdsLoading}
      error={error || crdsError}
      className={className}
      path={[{ label: "Flux Runtime" }]}
    >
      <FluxRuntimeComponent deployments={data?.deployments} crds={crds?.crds} />
    </Page>
  );
}

export default styled(FluxRuntime).attrs({ className: FluxRuntime.name })``;
