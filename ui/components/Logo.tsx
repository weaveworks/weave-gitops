import * as React from "react";
import { Link } from "react-router-dom";
import styled from "styled-components";
import images from "../lib/images";
import { V2Routes } from "../lib/types";
import Flex from "./Flex";

type Props = {
  className?: string;
};

function Logo({ className }: Props) {
  return (
    <Flex className={className} start>
      <Link to={V2Routes.Automations}>
        <img src={images.logoSrc} />
      </Link>
    </Flex>
  );
}

export default styled(Logo)`
  a {
    display: flex;
    align-items: center;
  }
  img {
    width: auto;
    height: 32px;
  }
  padding-left: ${(props) => props.theme.spacing.small};
  //this width plus small spacing (12px) lines up the breadcrumbs with the page title.
  width: 236px;
`;
