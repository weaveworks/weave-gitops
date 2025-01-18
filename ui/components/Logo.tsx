import * as React from "react";
import { Link } from "react-router";
import styled from "styled-components";
import { useInDarkMode } from "../hooks/theme";
import images from "../lib/images";
import { V2Routes } from "../lib/types";
import { Fade } from "../lib/utils";
import Flex from "./Flex";
import { topBarHeight } from "./Page";

type Props = {
  className?: string;
  collapsed: boolean;
  link?: string;
};

function Logo({ className, link = V2Routes.Automations, collapsed }: Props) {
  const dark = useInDarkMode();
  return (
    <Flex className={className} alignItems="center">
      <Link to={link}>
        <img src={dark ? images.logoDark : images.logoLight} />
      </Link>
      <Fade fade={collapsed}>
        <Link to={link}>
          <img src={dark ? images.logotypeLight : images.logotype} />
        </Link>
      </Fade>
    </Flex>
  );
}

export default styled(Logo)`
  a {
    display: flex;
    align-items: center;
  }
  img:first-child {
    margin-right: ${(props) => props.theme.spacing.xs};
  }
  img {
    width: auto;
    height: 32px;
  }
  padding-left: 16px;
  height: ${topBarHeight};
  width: ${(props) => (props.collapsed ? "36px" : "180px")};
  transition: width 0.5s;
`;
