import styled from "styled-components";
// eslint-disable-next-line
import { spacing } from "../typedefs/styled";

type Spacing = keyof typeof spacing;

type Props = {
  className?: string;
  margin?: Spacing;
  padding?: Spacing;
  m?: Spacing[];
};

const mFn = (n: number) => (props) =>
  props.m && props.theme.spacing[props.m[n]];

const Spacer = styled.div<Props>`
  margin: ${(props) => props.theme.spacing[props.margin]};
  padding: ${(props) => props.theme.spacing[props.padding]};
  margin-top: ${mFn(0)};
  margin-bottom: ${mFn(1)};
  margin-right: ${mFn(2)};
  margin-left: ${mFn(3)};
`;

export default Spacer;
