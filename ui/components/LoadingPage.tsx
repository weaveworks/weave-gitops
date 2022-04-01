import { CircularProgress } from "@material-ui/core";
import * as React from "react";
import styled from "styled-components";
import theme from "./../lib/theme";
import Flex from "./Flex";

type Props = {
  className?: string;
};

function LoadingPage({ className }: Props) {
  return (
    <Flex className={className} center wide align>
      <CircularProgress style={{ color: theme.colors.primary }} />
    </Flex>
  );
}

export default styled(LoadingPage)``;
