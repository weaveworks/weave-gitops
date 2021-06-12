import * as React from "react";
import styled from "styled-components";
/*eslint import/no-unresolved: [0]*/
// @ts-ignore
import logoSrc from "url:../images/logo.svg";
import Spacer from "./Spacer";

type Props = {
  className?: string;
};

function Logo({ className }: Props) {
  return (
    <div className={className}>
      <Spacer padding="medium">
        <img src={logoSrc} />
      </Spacer>
    </div>
  );
}

export default styled(Logo)``;
