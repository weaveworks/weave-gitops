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

export const Breadcrumbs = () => {
  const { currentPage } = useNavigation();
  const { search } = useLocation();

  const parentValue = getParentNavValue(currentPage) as V2Routes;
  const label = getPageLabel(parentValue);
  const parsed = qs.parse(search);
  const crumblings = currentPage.split("/");
  const crumbs = [];
  if (crumblings.length > 2) {
    for (let i = 2; i < crumblings.length; i++) crumbs.push(crumblings[i]);
  }
  if (parsed.name) crumbs.push(parsed.name);

  return (
    <Flex align>
      <Link
        to={parentValue || ""}
        textProps={{ bold: true, size: "large", color: "neutral40" }}
      >
        {label}
      </Link>
      {crumbs.map((crumb) => (
        <React.Fragment key={crumb}>
          <Icon
            type={IconType.NavigateNextIcon}
            size="large"
            color="neutral40"
          />
          <Link
            to={formatURL(currentPage, parsed)}
            textProps={{ size: "large", color: "neutral40" }}
          >
            {crumb}
          </Link>
        </React.Fragment>
      ))}
    </Flex>
  );
};

export default styled(Breadcrumbs)``;
