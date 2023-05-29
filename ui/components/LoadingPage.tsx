import { CircularProgress } from "@material-ui/core";
import * as React from "react";
import styled from "styled-components";
import Flex from "./Flex";

type Props = {
  className?: string;
};

function LoadingPage({ className }: Props) {
  return (
    <Flex className={className} center wide align>
      <CircularProgress color="primary" />
    </Flex>
  );
}

export default styled(LoadingPage)``;
