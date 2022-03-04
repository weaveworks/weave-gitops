import qs from "query-string";
import React from "react";
import { useLocation } from "react-router-dom";
import styled from "styled-components";
import useNavigation from "../hooks/navigation";
import { formatURL, getPageLabel, getParentNavValue } from "../lib/nav";
import { V2Routes } from "../lib/types";
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
  const { search } = useLocation();

  const parentValue = getParentNavValue(currentPage) as V2Routes;
  const label = getPageLabel(parentValue);
  const parsed = qs.parse(search);

  if (!parentValue) {
    throw new Error("invalid route");
  }

  return (
    <Flex align>
      <CrumbLink to={parentValue} textProps={{ bold: true }}>
        {label}
      </CrumbLink>

      {parentValue !== currentPage && (
        <>
          <Icon
            type={IconType.NavigateNextIcon}
            size="large"
            color="neutral00"
          />
          <CrumbLink to={formatURL(currentPage, parsed)}>
            {parsed.name}
          </CrumbLink>
        </>
      )}
    </Flex>
  );
};

export default styled(Breadcrumbs)``;
