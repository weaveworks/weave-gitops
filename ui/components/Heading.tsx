import * as React from "react";
import styled from "styled-components";

type Props = {
  className?: string;
  level: 1 | 2 | 3 | 4 | 5;
  children: any;
};

function Heading({ className, children, level }: Props) {
  return React.createElement(`h${level}`, { className }, children);
}

export default styled(Heading).attrs({ className: Heading.name })`
  font-size: 20px;
`;
