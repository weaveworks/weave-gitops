import styled from "styled-components";
import { colors, fontSizes } from "../typedefs/styled";

type Props = {
  className?: string;
  size?: keyof typeof fontSizes;
  bold?: boolean;
  semiBold?: boolean;
  capitalize?: boolean;
  italic?: boolean;
  color?: keyof typeof colors;
};

const Text = styled.span<Props>`
  font-family: ${(props) => props.theme.fontFamilies.regular};
  font-size: ${(props) => props.theme.fontSizes[props.size]};
  font-weight: ${(props) => {
    if (props.bold) return "800";
    else if (props.semiBold) return "600";
    else return "400";
  }};
  text-transform: ${(props) => (props.capitalize ? "uppercase" : "none")};
  font-style: ${(props) => (props.italic ? "italic" : "normal")};
  color: ${(props) => props.theme.colors[props.color as any]};
`;

Text.defaultProps = {
  size: "normal",
};

export default Text;
