import * as React from "react";
import {useContext} from "react";
import styled from "styled-components";
import FluxRuntimeComponent from "../../components/FluxRuntime";
import Page from "../../components/Page";
import {CoreClientContext} from "../../contexts/CoreClientContext";
import { useFeatureFlags } from "../../hooks/featureflags";
import {useListFluxCrds, useListFluxRuntimeObjects, useListRuntimeCrds, useListRuntimeObjects} from "../../hooks/flux";

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
