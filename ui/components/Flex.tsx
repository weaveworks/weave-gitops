import * as React from "react";
import styled from "styled-components";

type Props = {
  className?: string;
  column?: boolean;
  align?: boolean;
  height?: string;
  between?: boolean;
  center?: boolean;
  wide?: boolean;
  wrap?: boolean;
  shadow?: boolean;
  onMouseEnter?: React.ReactEventHandler;
  onMouseLeave?: React.ReactEventHandler;
};

const Styled = (component) => styled(component)`
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

class Flex extends React.PureComponent<Props> {
  render() {
    const { className, children, onMouseEnter, onMouseLeave } = this.props;
    return (
      <div
        className={className}
        onMouseEnter={onMouseEnter}
        onMouseLeave={onMouseLeave}
      >
        {children}
      </div>
    );
  }
}

export const MessageFlex = styled(Styled(Flex))`
  box-sizing: border-box;
  width: 560px;
  padding: ${({ theme }) => theme.spacing.medium}
    ${({ theme }) => theme.spacing.xl} ${({ theme }) => theme.spacing.xxl};
  border-radius: 10px;
  background-color: #ffffffd9;
  color: ${({ theme }) => theme.colors.neutral30};
`;

MessageFlex.defaultProps = {
  shadow: true,
  align: true,
  column: true,
};

export default Styled(Flex);
