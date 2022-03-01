import React from "react";
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
        {getPageLabel(currentPage)}
      </CrumbLink>
    </Flex>
  );
};

export default styled(Breadcrumbs)``;
