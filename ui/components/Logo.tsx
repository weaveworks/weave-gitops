import * as React from "react";
import { Link } from "react-router-dom";
import styled from "styled-components";
import images from "../lib/images";
import { V2Routes } from "../lib/types";
import Flex from "./Flex";

type Props = {
  className?: string;
  collapsed: boolean;
};

function Logo({ className }: Props) {
  return (
    <Flex className={className} wide>
      <Link to={V2Routes.Automations}>
        <img src={images.weaveG} />
      </Link>
      <img src={images.logotype} />
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
  img:nth-child(2) {
    opacity: ${(props) => (props.collapsed ? 0 : 1)};
    transition: opacity 200ms cubic-bezier(0.4, 0, 0.2, 1) 0ms;
  }
  img {
    width: auto;
    height: 32px;
  }
  padding-left: 20px;
  //nav width: 200px - space btwn nav and content: 24px - content padding: 24px. All together 248px - 20 for left padding to line up with detail page titles.
  width: ${(props) => (props.collapsed ? "92px" : "228px")};
  transition: width 0.5s;
`;
