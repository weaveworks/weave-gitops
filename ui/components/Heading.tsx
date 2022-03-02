import * as React from "react";
import styled from "styled-components";

type Props = {
  className?: string;
  level: 1 | 2 | 3 | 4 | 5;
  children: any;
};

function Heading({ className, children, level }: Props) {
  return React.createElement(
    `h${level}`,
    { className: `${className} h${level}` },
    children
  );
}

export default styled(Heading).attrs({ className: Heading.name })`
  &.h1 {
    font-weight: 600;
    font-size: 20px;
    line-height: 24px;
  }

  &.h2 {
    font-size: 20px;
    line-height: 24px;
    margin-top: 0;
    margin-bottom: 24px;
    font-weight: 400;
    color: #737373;
  }

  &.h3 {
    font-size: 14px;
    line-height: 17px;
  }
`;
