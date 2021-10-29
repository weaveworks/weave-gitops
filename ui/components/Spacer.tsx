import styled from "styled-components";
import { spacing } from "../typedefs/styled";

type Props = {
  className?: string;
  margin?: keyof typeof spacing;
  customMargin?: string;
  padding?: keyof typeof spacing;
  customPadding?: string;
};

const Spacer = styled.div<Props>`
  margin: ${(props) =>
    props.customMargin
      ? props.customMargin
      : props.theme.spacing[props.margin]};
  padding: ${(props) =>
    props.customPadding
      ? props.customPadding
      : props.theme.spacing[props.padding]};
`;

export default Spacer;
