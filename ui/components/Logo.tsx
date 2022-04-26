import * as React from "react";
import styled from "styled-components";
import images from "../lib/images";
import { V2Routes } from "../lib/types";
import Flex from "./Flex";
import Link from "./Link";
import Spacer from "./Spacer";

type Props = {
  className?: string;
};

function Logo({ className }: Props) {
  return (
    <Flex className={className} start>
      <Link to={V2Routes.Automations}>
        <Flex align>
          <img src={images.logoSrc} style={{ height: 56 }} />
          <Spacer padding="xxs" />
          <img src={images.titleSrc} />
        </Flex>
      </Link>
    </Flex>
  );
}

export default styled(Logo)`
  padding-left: ${(props) => props.theme.spacing.medium};
  //this width plus medium spacing (24px) lines up the breadcrumbs with the page title.
  width: 224px;
`;
