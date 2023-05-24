import { Tooltip } from "@material-ui/core";
import qs from "query-string";
import React from "react";
import { useLocation } from "react-router-dom";
import styled from "styled-components";
import useNavigation from "../hooks/navigation";
import { getPageLabel, getParentNavValue } from "../lib/nav";
import { V2Routes } from "../lib/types";
import Flex from "./Flex";
import Icon, { IconType } from "./Icon";
import Link from "./Link";
import Text from "./Text";

const EllipsedText = styled(Text)<{ maxWidth?: string }>`
  max-width: ${(prop) => prop.maxWidth || "300px"};
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
`;

export const Breadcrumbs = () => {
  const { currentPage } = useNavigation();
  const { search } = useLocation();
  const parentValue = getParentNavValue(currentPage) as V2Routes;
  const label = getPageLabel(parentValue);
  const parsed = qs.parse(search);
  return (
    <Flex align>
      <Link
        to={parentValue || ""}
        textProps={{ bold: true, size: "large", color: "neutral40" }}
      >
        {label}
      </Link>
      {parentValue !== currentPage && parsed.name && (
        <>
          {label && (
            <Icon
              type={IconType.NavigateNextIcon}
              size="large"
              color="neutral40"
            />
          )}
          <Tooltip title={parsed.name} placement="bottom">
            <EllipsedText size="large" color="neutral40" className="ellipsis">
              {parsed.name}
            </EllipsedText>
          </Tooltip>
        </>
      )}
    </Flex>
  );
};

export default styled(Breadcrumbs).attrs({ className: Breadcrumbs.name })``;
