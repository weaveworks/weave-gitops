import React from "react";
import { useResolvedPath } from "react-router-dom";
import ImageRepositoriesTable from "../../components/ImageAutomation/repositories/ImageRepositoriesTable";
import ImageAutomationUpdatesTable from "../../components/ImageAutomation/updates/ImageAutomationUpdatesTable";
import { routeTab } from "../../components/KustomizationDetail";
import SubRouterTabs, { RouterTab } from "../../components/SubRouterTabs";
import Flex from "../Flex";
import ImagePoliciesTable from "./policies/ImagePoliciesTable";

const ImageAutomation = () => {
  const path = useResolvedPath("").pathname;

  const tabs: Array<routeTab> = [
    {
      name: "Image Repositories",
      path: `${path}/repositories`,
      component: () => {
        return <ImageRepositoriesTable />;
      },
      visible: true,
    },
    {
      name: "Image Policies",
      path: `${path}/policies`,
      component: () => {
        return <ImagePoliciesTable />;
      },
      visible: true,
    },
    {
      name: "Image Update Automations",
      path: `${path}/updates`,
      component: () => {
        return <ImageAutomationUpdatesTable />;
      },
      visible: true,
    },
  ];
  return (
    <Flex wide tall column>
      <SubRouterTabs rootPath={tabs[0].path} clearQuery>
        {tabs.map(
          (subRoute, index) =>
            subRoute.visible && (
              <RouterTab name={subRoute.name} path={subRoute.path} key={index}>
                {subRoute.component()}
              </RouterTab>
            ),
        )}
      </SubRouterTabs>
    </Flex>
  );
};

export default ImageAutomation;
