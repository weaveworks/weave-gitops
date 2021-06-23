import { CircularProgress } from "@material-ui/core";
import * as React from "react";
import styled from "styled-components";
import Flex from "./Flex";

type Props = {
  className?: string;
};

function LoadingPage({ className }: Props) {
  return (
    <div>
      <Flex className={className} center wide align>
        <CircularProgress />
      </Flex>
    </div>
  );
}

export default styled(LoadingPage)``;
