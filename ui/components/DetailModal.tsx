import * as React from "react";
import styled from "styled-components";

type Props = {
  className?: string;
};

function DetailModal({ className }: Props) {
  return <div className={className}></div>;
}

export default styled(DetailModal).attrs({ className: DetailModal.name })``;
