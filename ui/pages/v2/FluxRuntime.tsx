import * as React from "react";
import {useContext} from "react";
import styled from "styled-components";
import FluxRuntimeComponent from "../../components/FluxRuntime";
import Page from "../../components/Page";
import {CoreClientContext} from "../../contexts/CoreClientContext";
import { useFeatureFlags } from "../../hooks/featureflags";
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
  const { featureFlags: flags } = useContext(CoreClientContext);
  const { isFlagEnabled } = useFeatureFlags();

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

export default styled(FluxRuntime).attrs({ className: FluxRuntime.name })``;
