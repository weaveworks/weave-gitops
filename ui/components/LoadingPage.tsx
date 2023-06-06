import { CircularProgress } from "@material-ui/core";
import * as React from "react";
import styled from "styled-components";
import Flex from "./Flex";

type Props = {
  className?: string;
};

const Box = styled(Flex)`
  margin-top: ${(props) => props.theme.spacing.base};
  margin-bottom: ${(props) => props.theme.spacing.base};
`;

function LoadingPage({ className }: Props) {
  return (
    <Box className={className} center wide align>
      <CircularProgress color="primary" />
    </Box>
  );
}

export default styled(LoadingPage)``;
