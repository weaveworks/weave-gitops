import AccountCircleIcon from "@material-ui/icons/AccountCircle";
import AddIcon from "@material-ui/icons/Add";
import CheckCircleIcon from "@material-ui/icons/CheckCircle";
import ArrowUpwardIcon from "@material-ui/icons/ArrowUpward";
import DeleteIcon from "@material-ui/icons/Delete";
import NavigateNextIcon from "@material-ui/icons/NavigateNext";
import SaveAltIcon from "@material-ui/icons/SaveAlt";
import ErrorIcon from "@material-ui/icons/Error";
import HourglassFullIcon from "@material-ui/icons/HourglassFull";

import LaunchIcon from "@material-ui/icons/Launch";
import * as React from "react";
import styled from "styled-components";
import { colors, spacing } from "../typedefs/styled";
import Flex from "./Flex";
import Text from "./Text";

type Props = {
  className?: string;
  type: string;
  color?: keyof typeof colors;
  text?: string;
  size: keyof typeof spacing;
};

function getIcon(type: string) {
  switch (type) {
    case "Checkmark":
      return CheckCircleIcon;

    case "Account":
      return AccountCircleIcon;

    case "ExternaTtab":
      return LaunchIcon;

    case "Add":
      return AddIcon;

    case "ArrowUpward":
      return ArrowUpwardIcon;

    case "Delete":
      return DeleteIcon;

    case "NavigateNext":
      return NavigateNextIcon;

    case "SaveAlt":
      return SaveAltIcon;

    case "Current":
      return CheckCircleIcon;

    case "InProgress":
      return HourglassFullIcon;

    case "Failed":
      return ErrorIcon;

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

  ${Text} {
    margin-left: 4px;
    color: ${(props) => props.theme.colors[props.color as any]};
  }
`;
