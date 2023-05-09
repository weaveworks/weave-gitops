import * as React from "react";
import { Link } from "react-router-dom";
import styled from "styled-components";
import images from "../lib/images";
import { V2Routes } from "../lib/types";
import { Fade } from "../lib/utils";
import Flex from "./Flex";

type Props = {
  className?: string;
  collapsed: boolean;
  link?: string;
};

function Logo({ className, link, collapsed }: Props) {
  return (
    <Flex className={className} wide>
      <Link to={link || V2Routes.Automations}>
        <img src={images.logoLight} />
      </Link>
      <Fade fade={collapsed}>
        <Link to={link || V2Routes.Automations}>
          <img src={images.logotype} />
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
  padding-left: 20px;
  //nav width: 200px - space btwn nav and content: 24px - content padding: 24px. All together 248px - 20 for left padding to line up with detail page titles.
  width: ${(props) => (props.collapsed ? "92px" : "228px")};
  transition: width 0.5s;
`;
