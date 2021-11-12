import AccountCircleIcon from "@material-ui/icons/AccountCircle";
import AddIcon from "@material-ui/icons/Add";
import ArrowDownwardIcon from "@material-ui/icons/ArrowDownward";
import CheckCircleIcon from "@material-ui/icons/CheckCircle";
import DeleteForeverIcon from "@material-ui/icons/DeleteForever";
import LaunchIcon from "@material-ui/icons/Launch";
import * as React from "react";
import styled from "styled-components";
import { colors, spacing } from "../typedefs/styled";
import Flex from "./Flex";
import Text from "./Text";

export enum IconType {
  CheckMark,
  Account,
  ExternalTab,
  Add,
  ArrowDownward,
  DeleteForever,
}
type Props = {
  className?: string;
  type: IconType;
  color?: keyof typeof colors;
  text?: string;
  size: keyof typeof spacing;
};

function getIcon(i: IconType) {
  switch (i) {
    case IconType.CheckMark:
      return CheckCircleIcon;

    case IconType.Account:
      return AccountCircleIcon;

    case IconType.ExternalTab:
      return LaunchIcon;

    case IconType.Add:
      return AddIcon;

    case IconType.ArrowDownward:
      return ArrowDownwardIcon;

    case IconType.DeleteForever:
      return DeleteForeverIcon;

    default:
      break;
  }
}

function Icon({ className, type, text, color }: Props) {
  return (
    <Flex align className={className}>
      {React.createElement(getIcon(type))}
      {text && (
        <Text color={color} bold>
          {text}
        </Text>
      )}
    </Flex>
  );
}

export default styled(Icon)`
  svg {
    fill: ${(props) => props.theme.colors[props.color as any]} !important;
    height: ${(props) => props.theme.spacing[props.size as any]};
    width: ${(props) => props.theme.spacing[props.size as any]};
  }
  &.upward {
    transform: rotate(180deg);
  }
  &.downward {
    transform: initial;
  }

  ${Text} {
    margin-left: 4px;
    color: ${(props) => props.theme.colors[props.color as any]};
  }
`;
