import AddIcon from "@material-ui/icons/Add";
import ArrowDropDownIcon from "@material-ui/icons/ArrowDropDown";
import ArrowUpwardIcon from "@material-ui/icons/ArrowUpward";
import CheckCircleIcon from "@material-ui/icons/CheckCircle";
import ClearIcon from "@material-ui/icons/Clear";
import DeleteIcon from "@material-ui/icons/Delete";
import ErrorIcon from "@material-ui/icons/Error";
import LogoutIcon from "@material-ui/icons/ExitToApp";
import FileCopyIcon from "@material-ui/icons/FileCopyOutlined";
import FilterIcon from "@material-ui/icons/FilterList";
import HourglassFullIcon from "@material-ui/icons/HourglassFull";
import LaunchIcon from "@material-ui/icons/Launch";
import NavigateBeforeIcon from "@material-ui/icons/NavigateBefore";
import NavigateNextIcon from "@material-ui/icons/NavigateNext";
import PauseIcon from "@material-ui/icons/Pause";
import PersonIcon from "@material-ui/icons/Person";
import PlayIcon from "@material-ui/icons/PlayArrow";
import RemoveCircleIcon from "@material-ui/icons/RemoveCircle";
import SaveAltIcon from "@material-ui/icons/SaveAlt";
import SearchIcon from "@material-ui/icons/Search";
import SkipNextIcon from "@material-ui/icons/SkipNext";
import SkipPreviousIcon from "@material-ui/icons/SkipPrevious";
import * as React from "react";
import styled from "styled-components";
import images from "../lib/images";
import DocsIcon from "./NavIcons/DocsIcon";
// eslint-disable-next-line
import { colors, spacing } from "../typedefs/styled";
import Flex from "./Flex";
import ApplicationsIcon from "./NavIcons/ApplicationsIcon";
import ClustersIcon from "./NavIcons/ClustersIcon";
import DeliveryIcon from "./NavIcons/DeliveryIcon";
import ExploreIcon from "./NavIcons/ExploreIcon";
import FluxIcon from "./NavIcons/FluxIcon";
import GitOpsRunIcon from "./NavIcons/GitOpsRunIcon";
import GitOpsSetsIcon from "./NavIcons/GitOpsSetsIcon";
import ImageAutomationIcon from "./NavIcons/ImageAutomationIcon";
import NotificationsIcon from "./NavIcons/NotificationsIcon";
import PipelinesIcon from "./NavIcons/PipelinesIcon";
import PoliciesIcon from "./NavIcons/PoliciesIcon";
import PolicyConfigsIcon from "./NavIcons/PolicyConfigsIcon";
import SecretsIcon from "./NavIcons/SecretsIcon";
import SourcesIcon from "./NavIcons/SourcesIcon";
import TemplatesIcon from "./NavIcons/TemplatesIcon";
import TerraformIcon from "./NavIcons/TerraformIcon";
import WorkspacesIcon from "./NavIcons/WorkspacesIcon";
import Text from "./Text";

export enum IconType {
  CheckMark,
  Account,
  ExternalTab,
  AddIcon,
  ArrowUpwardIcon,
  ArrowDropDownIcon,
  DeleteIcon,
  SaveAltIcon,
  ErrorIcon,
  CheckCircleIcon,
  HourglassFullIcon,
  NavigateNextIcon,
  NavigateBeforeIcon,
  SkipNextIcon,
  SkipPreviousIcon,
  RemoveCircleIcon,
  FilterIcon,
  ClearIcon,
  SearchIcon,
  LogoutIcon,
  SuccessIcon,
  FailedIcon,
  SuspendedIcon,
  FileCopyIcon,
  ReconcileIcon,
  FluxIcon,
  FluxIconHover,
  DocsIcon,
  ApplicationsIcon,
  PlayIcon,
  PauseIcon,
  NotificationsIcon,
  SourcesIcon,
  ImageAutomationIcon,
  DeliveryIcon,
  GitOpsRunIcon,
  PipelinesIcon,
  TerraformIcon,
  GitOpsSetsIcon,
  PoliciesIcon,
  PolicyConfigsIcon,
  WorkspacesIcon,
  SecretsIcon,
  TemplatesIcon,
  ClustersIcon,
  ExploreIcon,
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
      return PersonIcon;

    case IconType.ExternalTab:
      return LaunchIcon;

    case IconType.AddIcon:
      return AddIcon;

    case IconType.ArrowUpwardIcon:
      return ArrowUpwardIcon;

    case IconType.DeleteIcon:
      return DeleteIcon;

    case IconType.SaveAltIcon:
      return SaveAltIcon;

    case IconType.CheckCircleIcon:
      return CheckCircleIcon;

    case IconType.HourglassFullIcon:
      return HourglassFullIcon;

    case IconType.ErrorIcon:
      return ErrorIcon;

    case IconType.NavigateNextIcon:
      return NavigateNextIcon;

    case IconType.NavigateBeforeIcon:
      return NavigateBeforeIcon;

    case IconType.SkipNextIcon:
      return SkipNextIcon;

    case IconType.SkipPreviousIcon:
      return SkipPreviousIcon;

    case IconType.RemoveCircleIcon:
      return RemoveCircleIcon;

    case IconType.FilterIcon:
      return FilterIcon;

    case IconType.ClearIcon:
      return ClearIcon;

    case IconType.SearchIcon:
      return SearchIcon;

    case IconType.LogoutIcon:
      return LogoutIcon;

    case IconType.SuccessIcon:
      return () => <img src={images.successSrc} />;

    case IconType.FailedIcon:
      return ErrorIcon;

    case IconType.SuspendedIcon:
      return () => <img src={images.suspendedSrc} />;

    case IconType.ReconcileIcon:
      return () => <img src={images.reconcileSrc} />;

    case IconType.ArrowDropDownIcon:
      return ArrowDropDownIcon;

    case IconType.FileCopyIcon:
      return FileCopyIcon;

    case IconType.PlayIcon:
      return PlayIcon;

    case IconType.PauseIcon:
      return PauseIcon;

    case IconType.SourcesIcon:
      return SourcesIcon;

    case IconType.ImageAutomationIcon:
      return ImageAutomationIcon;

    case IconType.DeliveryIcon:
      return DeliveryIcon;

    case IconType.GitOpsRunIcon:
      return GitOpsRunIcon;

    case IconType.PipelinesIcon:
      return PipelinesIcon;

    case IconType.TerraformIcon:
      return TerraformIcon;

    case IconType.ApplicationsIcon:
      return ApplicationsIcon;

    case IconType.DocsIcon:
      return DocsIcon;

    case IconType.FluxIcon:
      return FluxIcon;

    case IconType.GitOpsSetsIcon:
      return GitOpsSetsIcon;

    case IconType.NotificationsIcon:
      return NotificationsIcon;

    case IconType.PoliciesIcon:
      return PoliciesIcon;

    case IconType.PolicyConfigsIcon:
      return PolicyConfigsIcon;

    case IconType.SecretsIcon:
      return SecretsIcon;

    case IconType.TemplatesIcon:
      return TemplatesIcon;

    case IconType.WorkspacesIcon:
      return WorkspacesIcon;

    case IconType.ClustersIcon:
      return ClustersIcon;

    case IconType.ExploreIcon:
      return ExploreIcon;

    default:
      break;
  }
}

function Icon({ className, type, text, color }: Props) {
  return (
    <Flex align className={className}>
      {React.createElement(getIcon(type) || "span")}
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
    fill: ${(props) => props.theme.colors[props.color as any]};
    height: ${(props) => props.theme.spacing[props.size as any]};
    width: ${(props) => props.theme.spacing[props.size as any]};
    path,
    line,
    polygon,
    rect,
    circle,
    polyline {
      &.path-fill {
        fill: ${(props) => props.theme.colors[props.color as any]} !important;
        transition: fill 200ms cubic-bezier(0.4, 0, 0.2, 1) 0ms;
      }
      &.stroke-fill {
        stroke: ${(props) => props.theme.colors[props.color as any]} !important;
        transition: stroke 200ms cubic-bezier(0.4, 0, 0.2, 1) 0ms;
      }
    }
    rect {
      &.rect-height {
        height: ${(props) => props.theme.spacing[props.size as any]};
        width: ${(props) => props.theme.spacing[props.size as any]};
      }
    }
  }
  &.downward {
    transform: rotate(180deg);
  }
  &.upward {
    transform: initial;
  }
  ${Text} {
    margin-left: 4px;
    color: ${(props) => props.theme.colors[props.color as any]};
  }

  img {
    width: ${(props) => props.theme.spacing[props.size as any]};
  }
`;
