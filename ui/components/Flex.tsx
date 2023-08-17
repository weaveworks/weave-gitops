import * as React from "react";
import styled from "styled-components";

type Props = {
  className?: string;
  column?: boolean;
  align?: boolean;
  alignItems?: string;
  height?: string;
  between?: boolean;
  center?: boolean;
  wide?: boolean;
  wrap?: boolean;
  shadow?: boolean;
  tall?: string;
  start?: string;
  end?: string;
  gap?: string;
  onMouseEnter?: React.ReactEventHandler;
  onMouseLeave?: React.ReactEventHandler;
  "data-testid"?: string;
};

const Styled = (component) => styled(component)`
  display: flex;
  flex-direction: ${({ column }) => (column ? "column" : "row")};
  align-items: ${({ align, alignItems }) =>
    alignItems ? alignItems : align ? "center" : "start"};
  ${({ gap }) => gap && `gap: ${gap}px`};
  ${({ tall }) => tall && `height: 100%`};
  ${({ wide }) => wide && "width: 100%"};
  ${({ wrap }) => wrap && "flex-wrap: wrap"};
  ${({ start }) => start && "justify-content: flex-start"};
  ${({ end }) => end && "justify-content: flex-end"};
  ${({ between }) => between && "justify-content: space-between"};
  ${({ center }) => center && "justify-content: center"};
  ${({ shadow }) => shadow && "box-shadow: 5px 10px 50px 3px #0000001a"};
`;

class Flex extends React.PureComponent<Props> {
  render() {
    const {
      className,
      onMouseEnter,
      onMouseLeave,
      "data-testid": dataTestId,
      children,
    } = this.props;
    return (
      <div
        data-testid={dataTestId}
        className={className}
        onMouseEnter={onMouseEnter}
        onMouseLeave={onMouseLeave}
      >
        {children}
      </div>
    );
  }
}

export const ColumnOnBreakpoint = styled(Styled(Flex))<{ breakpoint: number }>`
  @media screen and (max-width: ${(props) => props.breakpoint}px) {
    flex-direction: column;
  }
`;

export default Styled(Flex);
