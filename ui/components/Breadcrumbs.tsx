import qs from "query-string";
import React from "react";
import { useLocation } from "react-router-dom";
import styled from "styled-components";
import useNavigation from "../hooks/navigation";
import { getPageLabel, getParentNavValue, V2Routes } from "../lib/nav";
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
  const parsed = qs.parse(useLocation().search);

  return (
    <Flex align>
      {parentValue !== currentPage && (
        <CrumbLink
          to={V2Routes[parentValue as V2Routes] || ""}
          textProps={{ bold: true }}
        >
          {getPageLabel(parentValue as V2Routes)}
        </CrumbLink>
      )}
      {parentValue !== currentPage && (
        <Icon type={IconType.NavigateNextIcon} size="large" color="neutral00" />
      )}
      <CrumbLink
        to={currentPage}
        textProps={parentValue === currentPage && { bold: true }}
      >
        {getPageLabel(currentPage, parsed && (parsed.name as string))}
      </CrumbLink>
    </Flex>
  );
};

export default styled(Breadcrumbs)``;
