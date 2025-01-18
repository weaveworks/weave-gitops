import * as React from "react";
import styled from "styled-components";
import Flex from "../components/Flex";
import { Crd, Deployment } from "../lib/api/core/types.pb";
import ControllersTable from "./ControllersTable";
import CrdsTable from "./CrdsTable";
import FluxVersionsTable, { FluxVersion } from "./FluxVersionsTable";
import { routeTab } from "./KustomizationDetail";
import SubRouterTabs, { RouterTab } from "./SubRouterTabs";

type Props = {
  className?: string;
  deployments?: Deployment[];
  crds?: Crd[];
};

const fluxVersionLabel = "app.kubernetes.io/version";
const partOfLabel = "app.kubernetes.io/part-of";
const fluxLabel = "flux";

function FluxRuntime({ className, deployments, crds }: Props) {
  const fluxVersions: { [key: string]: FluxVersion } = {};
  deployments
    .filter((d) => d.labels[partOfLabel] == fluxLabel)
    .forEach((d) => {
      const fv = d.labels[fluxVersionLabel];
      const k = `${fv}${d.clusterName}${d.namespace}`;
      if (!fluxVersions[k]) {
        fluxVersions[k] = {
          version: fv,
          clusterName: d.clusterName,
          namespace: d.namespace,
        };
      }
    });

  const tabs: Array<routeTab> = [
    {
      name: "Flux Versions",
      path: "flux",
      component: () => {
        return <FluxVersionsTable versions={Object.values(fluxVersions)} />;
      },
      visible: true,
    },
    {
      name: "Controllers",
      path: "controllers",
      component: () => {
        return <ControllersTable controllers={deployments} />;
      },
      visible: true,
    },
    {
      name: "CRDs",
      path: "crds",
      component: () => {
        return <CrdsTable crds={crds} />;
      },
      visible: true,
    },
  ];

  return (
    <Flex wide tall column className={className}>
      <>
        <SubRouterTabs rootPath={tabs[0].path} clearQuery>
          {tabs.map(
            (subRoute, index) =>
              subRoute.visible && (
                <RouterTab
                  name={subRoute.name}
                  path={subRoute.path}
                  key={index}
                >
                  {subRoute.component()}
                </RouterTab>
              ),
          )}
        </SubRouterTabs>
      </>
    </Flex>
  );
}

export default styled(FluxRuntime).attrs({ className: FluxRuntime.name })``;
