import { Tooltip } from "@material-ui/core";
import * as React from "react";
import styled from "styled-components";
import Flex from "./Flex";
import Text from "./Text";

type Props = {
  className?: string;
  title: string;
  info?: string;
  component?: JSX.Element;
};

function LargeInfo({ className, title, info, component }: Props) {
  return (
    <Flex gap="4" alignItems="baseline" className={className}>
      <Text capitalize semiBold color="neutral30">
        {title}:
      </Text>
      <Tooltip title={info || ""} placement="top">
        <Text size="large" color="neutral40">
          {component ? component : info || ""}
        </Text>
      </Tooltip>
    </Flex>
  );
}

export default styled(LargeInfo).attrs({
  className: LargeInfo.name,
})`
  margin: 0 ${(props) => props.theme.spacing.small};
  ${Text} {
    max-width: 150px;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }
`;
