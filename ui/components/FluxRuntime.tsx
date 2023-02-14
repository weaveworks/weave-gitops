import * as React from "react";
import { useRouteMatch } from "react-router-dom";
import styled from "styled-components";
import Flex from "../components/Flex";
import { Crd, Deployment } from "../lib/api/core/types.pb";
import ControllersTable from "./ControllersTable";
import CrdsTable from "./CrdsTable";
import FluxVersionsTable, { FluxVersion } from "./FluxVersionsTable";
import { routeTab } from "./KustomizationDetail";
import SubRouterTabs, { RouterTab } from "./SubRouterTabs";
import Text from "./Text";

const FluxVersionText = styled(Text)`
  font-weight: 700;
  margin-bottom: ${(props) => props.theme.spacing.medium};

  span {
    color: ${(props) => props.theme.colors.neutral40};
    font-weight: 400;
    margin-left: ${(props) => props.theme.spacing.xs};
  }
`;

type Props = {
  className?: string;
  deployments?: Deployment[];
  crds?: Crd[];
};
const fluxVersionLabel = "app.kubernetes.io/version";

function FluxRuntime({ className, deployments, crds }: Props) {
  const { path } = useRouteMatch();
  const tabs: Array<routeTab> = [
    {
      name: "Controllers",
      path: `${path}/controllers`,
      component: () => {
        return <ControllersTable controllers={deployments} />;
      },
      visible: true,
    },
    {
      name: "CRDs",
      path: `${path}/crds`,
      component: () => {
        return <CrdsTable crds={crds} />;
      },
      visible: true,
    },
  ];
  const fluxVersions: { [key: string]: FluxVersion } = {};
  deployments.forEach((d) => {
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

  const supportMultipleFlux =
    Object.keys(fluxVersions).length > 1 ? true : false;

  if (supportMultipleFlux) {
    tabs.unshift({
      name: "Flux Versions",
      path: `${path}/flux`,
      component: () => {
        return <FluxVersionsTable versions={Object.values(fluxVersions)} />;
      },
      visible: true,
    });
  }
  return (
    <Flex wide tall column className={className}>
      <>
        {!supportMultipleFlux && deployments[0]?.labels[fluxVersionLabel] && (
          <FluxVersionText color="neutral30" titleHeight={true}>
            This cluster is running Flux version:
            <span>{deployments[0].labels[fluxVersionLabel]}</span>
          </FluxVersionText>
        )}
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
              )
          )}
        </SubRouterTabs>
      </>
    </Flex>
  );
}

export default styled(FluxRuntime).attrs({ className: FluxRuntime.name })``;
