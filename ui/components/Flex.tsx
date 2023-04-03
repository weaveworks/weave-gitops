import React from "react";
import styled from "styled-components";

type Props = {
  align?: boolean;
  between?: boolean;
  center?: boolean;
  column?: boolean;
  end?: boolean;
  shadow?: boolean;
  start?: boolean;
  tall?: boolean;
  wide?: boolean;
  wrap?: boolean;
} & React.HTMLAttributes<HTMLDivElement>;

const Styled = (component: React.ComponentType<Props>) => styled(
  component
)<Props>`
  display: flex;
  flex-direction: ${({ column }) => (column ? "column" : "row")};
  align-items: ${({ align }) => (align ? "center" : "start")};
  ${({ tall }) => tall && `height: 100%`};
  ${({ wide }) => wide && "width: 100%"};
  ${({ wrap }) => wrap && "flex-wrap: wrap"};
  ${({ start }) => start && "justify-content: flex-start"};
  ${({ end }) => end && "justify-content: flex-end"};
  ${({ between }) => between && "justify-content: space-between"};
  ${({ center }) => center && "justify-content: center"};
  ${({ shadow }) => shadow && "box-shadow: 5px 10px 50px 3px #0000001a"};
`;

function Flex(props: Props) {
  const {
    // don't forward to div
    align,
    between,
    center,
    column,
    end,
    shadow,
    start,
    tall,
    wide,
    wrap,
    // forward to div
    children,
    ...rest
  } = props;
  return <div {...rest}>{children}</div>;
}

export default Styled(Flex);
