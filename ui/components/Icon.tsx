import {
  DashboardOutlined,
  DnsOutlined,
  PolicyOutlined,
  TabOutlined,
  VerifiedUserOutlined,
  VpnKeyOutlined,
} from "@material-ui/icons";
import AddIcon from "@material-ui/icons/Add";
import ArrowDropDownIcon from "@material-ui/icons/ArrowDropDown";
import ArrowUpwardIcon from "@material-ui/icons/ArrowUpward";
import DocsIcon from "@material-ui/icons/AssignmentOutlined";
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
import NotificationsBellIcon from "@material-ui/icons/NotificationsNone";
import PauseIcon from "@material-ui/icons/Pause";
import PersonIcon from "@material-ui/icons/Person";
import PlayIcon from "@material-ui/icons/PlayArrow";
import RemoveCircleIcon from "@material-ui/icons/RemoveCircle";
import SaveAltIcon from "@material-ui/icons/SaveAlt";
import SearchIcon from "@material-ui/icons/Search";
import ApplicationsIcon from "@material-ui/icons/SettingsApplicationsOutlined";
import SkipNextIcon from "@material-ui/icons/SkipNext";
import SkipPreviousIcon from "@material-ui/icons/SkipPrevious";
import * as React from "react";
import styled from "styled-components";
import images from "../lib/images";
// eslint-disable-next-line
import { colors, spacing } from "../typedefs/styled";
import DeliveryIcon from "./DeliveryIcon";
import Flex from "./Flex";
import GitOpsRunIcon from "./GitOpsRunIcon";
import ImageAutomationIcon from "./ImageAutomationIcon";
import PipelinesIcon from "./PipelinesIcon";
import SourcesIcon from "./SourcesIcon";
import TerraformIcon from "./TerraformIcon";
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
  NotificationsBell,
  SourcesIcon,
  ImageAutomationIcon,
  DeliveryIcon,
  GitOpsRunIcon,
  PipelinesIcon,
  TerraformIcon,
  DashboardOutlined,
  DnsOutlined,
  PolicyOutlined,
  TabOutlined,
  VerifiedUserOutlined,
  VpnKeyOutlined,
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

    case IconType.ApplicationsIcon:
      return ApplicationsIcon;

    case IconType.DocsIcon:
      return DocsIcon;

    case IconType.FluxIcon:
      return () => <img src={images.fluxIconSrc} />;

    case IconType.FluxIconHover:
      return () => <img src={images.fluxIconHoverSrc} />;

    case IconType.PlayIcon:
      return PlayIcon;

    case IconType.PauseIcon:
      return PauseIcon;

    case IconType.NotificationsBell:
      return NotificationsBellIcon;

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

    case IconType.DashboardOutlined:
      return DashboardOutlined;

    case IconType.DnsOutlined:
      return DnsOutlined;

    case IconType.PolicyOutlined:
      return PolicyOutlined;

    case IconType.TabOutlined:
      return TabOutlined;

    case IconType.VerifiedUserOutlined:
      return VerifiedUserOutlined;

    case IconType.VpnKeyOutlined:
      return VpnKeyOutlined;

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

    path {
      &.image-automation,
      &.sources,
      &.delivery,
      &.pipelines,
      &.run-icon,
      &.terraform-icon {
        fill: ${(props) => props.theme.colors[props.color as any]} !important;
        transition: fill 200ms cubic-bezier(0.4, 0, 0.2, 1) 0ms;
      }
    }
    &.sources {
      fill: none !important;
      rect {
        stroke: ${(props) => props.theme.colors[props.color as any]} !important;
        transition: stroke 200ms cubic-bezier(0.4, 0, 0.2, 1) 0ms;
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
