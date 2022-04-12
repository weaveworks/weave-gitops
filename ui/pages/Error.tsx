// @ts-nocheck
import * as React from "react";
import styled from "styled-components";
import Lottie from "react-lottie-player";
import Content from "../components/Page";
import error404 from "../images/error404.json";

type Props = {
  className?: string;
  code?: string;
};

function ErrorPage({ className }: Props) {
  return (
    <Content>
      <Lottie loop animationData={error404} play style={{ height: 650 }} />
    </Content>
  );
}

export default styled(ErrorPage)``;
