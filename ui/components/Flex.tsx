import * as React from "react";
import styled from "styled-components";

type Props = {
  className?: string;
  column?: boolean;
  align?: boolean;
  between?: boolean;
  center?: boolean;
  wide?: boolean;
  wrap?: boolean;
};

const Styled = (component) => styled(component)`
  display: flex;
  flex-direction: ${(props) => (props.column ? "column" : "row")};
  align-items: ${(props) => (props.align ? "center" : "start")};
  ${({ between }) => between && "justify-content: space-between"};
  ${({ center }) => center && "justify-content: center"};
  ${({ wide }) => wide && "width: 100%"};
  ${({ wrap }) => wrap && "flex-wrap: wrap"};
  ${({ end }) => end && "justify-content: flex-end"};
`;

class Flex extends React.PureComponent<Props> {
  render() {
    const { className, children } = this.props;
    return <div className={className}>{children}</div>;
  }
}

export default Styled(Flex);
