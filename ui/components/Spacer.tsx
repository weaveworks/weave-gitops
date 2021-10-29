import styled from "styled-components";
import { spacing } from "../typedefs/styled";

type Props = {
  className?: string;
  margin?: keyof typeof spacing;
  padding?: keyof typeof spacing;
};

const Spacer = styled.div<Props>`
  margin: ${(props) => props.theme.spacing[props.margin]};
  padding: ${(props) => props.theme.spacing[props.padding]};
`;

export default Spacer;
