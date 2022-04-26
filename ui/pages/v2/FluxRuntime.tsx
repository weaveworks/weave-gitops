import * as React from "react";
import styled from "styled-components";
import FluxRuntimeComponent from "../../components/FluxRuntime";
import Page from "../../components/Page";
import { useListFluxRuntimeObjects } from "../../hooks/flux";

type Props = {
  className?: string;
};

function FluxRuntime({ className }: Props) {
  const { data, isLoading, error } = useListFluxRuntimeObjects();

  return (
    <Page loading={isLoading} error={error} className={className}>
      <FluxRuntimeComponent deployments={data?.deployments} />
    </Page>
  );
}

export default styled(FluxRuntime).attrs({ className: FluxRuntime.name })``;
