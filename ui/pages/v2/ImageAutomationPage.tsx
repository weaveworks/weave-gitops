import * as React from "react";
import { useRouteMatch } from "react-router-dom";
import styled from "styled-components";
import Flex from "../../components/Flex";
import ImageAutomationsTable from "../../components/ImageAutomation/ImageAutomationsTable";
import ImageRepositoriesTable from "../../components/ImageAutomation/ImageRepositoriesTable";
import { routeTab } from "../../components/KustomizationDetail";
import Page from "../../components/Page";
import SubRouterTabs, { RouterTab } from "../../components/SubRouterTabs";

type Props = {
  className?: string;
};

function ImageAutomationPage({ className }: Props) {
  const { path } = useRouteMatch();
  const tabs: Array<routeTab> = [
    {
      name: "Image Update Automations",
      path: `${path}/updates`,
      component: () => {
        return <ImageAutomationsTable />;
      },
      visible: true,
    },
    {
      name: "Image Repositories",
      path: `${path}/repositories`,
      component: () => {
        return <ImageRepositoriesTable />;
      },
      visible: true,
    },
  ];
  return (
    <Page>
      <Flex wide tall column className={className}>
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
      </Flex>
    </Page>
  );
}

export default styled(ImageAutomationPage)``;
