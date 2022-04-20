import * as React from "react";
import styled from "styled-components";
import Lottie from "react-lottie-player";
import Content from "../components/Page";
import error404 from "../images/error404.json";
import Flex from "../components/Flex";

type Props = {
  className?: string;
  code?: string;
};

function ErrorPage({ className }: Props) {
  return (
    <Content>
      <Flex center wide align>
        <Lottie loop animationData={error404} play style={{ height: 650 }} />
      </Flex>
    </Content>
  );
}

export default styled(ErrorPage)``;
