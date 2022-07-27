import * as React from "react";
import { useRouteMatch } from "react-router-dom";
import styled from "styled-components";
import Flex from "../components/Flex";
import { Crd, Deployment } from "../lib/api/core/types.pb";
import ControllersTable from "./ControllersTable";
import CrdsTable from "./CrdsTable";
import SubRouterTabs, { RouterTab } from "./SubRouterTabs";

type Props = {
  className?: string;
  deployments?: Deployment[];
  crds?: Crd[];
};

function FluxRuntime({ className, deployments, crds }: Props) {
  const { path } = useRouteMatch();

  return (
    <Flex wide tall column className={className}>
      <SubRouterTabs rootPath={`${path}/controllers`} clearQuery>
        <RouterTab name="Controllers" path={`${path}/controllers`}>
          <ControllersTable controllers={deployments} />
        </RouterTab>
        <RouterTab name="CRDs" path={`${path}/crds`}>
          <CrdsTable crds={crds} />
        </RouterTab>
      </SubRouterTabs>
    </Flex>
  );
}

export default styled(FluxRuntime).attrs({ className: FluxRuntime.name })``;
