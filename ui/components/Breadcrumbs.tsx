import qs from "query-string";
import React from "react";
import styled from "styled-components";
import useNavigation from "../hooks/navigation";
import { getParentNavValue, V2Routes } from "../lib/nav";
import Flex from "./Flex";
import Icon, { IconType } from "./Icon";
import Link from "./Link";
import Text from "./Text";

const CrumbLink = styled(Link)`
  ${Text} {
    font-size: ${(props) => props.theme.fontSizes.large};
    color: ${(props) => props.theme.colors.neutral00};
  }
`;

export const Breadcrumbs = () => {
  const { currentPage } = useNavigation();
  const parentValue = getParentNavValue(currentPage);

  function pageLookup(route) {
    const parsed = qs.parse(location.search);
    if (route === V2Routes.Automations) return "Applications";
    else if (route === V2Routes.Sources) return "Sources";
    else if (route === V2Routes.FluxRuntime) return "Flux Runtime";
    else return parsed.name;
  }

  return (
    <Flex align>
      {parentValue !== currentPage && (
        <CrumbLink
          to={V2Routes[parentValue as V2Routes] || ""}
          textProps={{ bold: true }}
        >
          {pageLookup(parentValue)}
        </CrumbLink>
      )}
      {parentValue !== currentPage && (
        <Icon type={IconType.NavigateNextIcon} size="large" color="neutral00" />
      )}
      <CrumbLink
        to={currentPage}
        textProps={parentValue === currentPage && { bold: true }}
      >
        {pageLookup(currentPage)}
      </CrumbLink>
    </Flex>
  );
};

export default styled(Breadcrumbs)``;
