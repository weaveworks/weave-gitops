import AddIcon from "@mui/icons-material/Add";
import ArrowDownwardRoundedIcon from "@mui/icons-material/ArrowDownwardRounded";
import ArrowDropDownIcon from "@mui/icons-material/ArrowDropDown";
import ArrowUpwardIcon from "@mui/icons-material/ArrowUpward";
import ArrowUpwardRoundedIcon from "@mui/icons-material/ArrowUpwardRounded";
import CallMade from "@mui/icons-material/CallMade";
import CallReceived from "@mui/icons-material/CallReceived";
import CheckCircleIcon from "@mui/icons-material/CheckCircle";
import ClearIcon from "@mui/icons-material/Clear";
import DeleteIcon from "@mui/icons-material/Delete";
import EditIcon from "@mui/icons-material/Edit";
import ErrorIcon from "@mui/icons-material/Error";
import LogoutIcon from "@mui/icons-material/ExitToApp";
import FileCopyIcon from "@mui/icons-material/FileCopyOutlined";
import FilterIcon from "@mui/icons-material/FilterList";
import FindInPage from "@mui/icons-material/FindInPage";
import HourglassFullIcon from "@mui/icons-material/HourglassFull";
import InfoIcon from "@mui/icons-material/Info";
import KeyboardArrowDownIcon from "@mui/icons-material/KeyboardArrowDown";
import KeyboardArrowRightIcon from "@mui/icons-material/KeyboardArrowRight";
import LaunchIcon from "@mui/icons-material/Launch";
import NavigateBeforeIcon from "@mui/icons-material/NavigateBefore";
import NavigateNextIcon from "@mui/icons-material/NavigateNext";
import PauseIcon from "@mui/icons-material/Pause";
import PersonIcon from "@mui/icons-material/Person";
import PlayIcon from "@mui/icons-material/PlayArrow";
import Remove from "@mui/icons-material/Remove";
import RemoveCircleIcon from "@mui/icons-material/RemoveCircle";
import SaveAltIcon from "@mui/icons-material/SaveAlt";
import SearchIcon from "@mui/icons-material/Search";
import SettingsIcon from "@mui/icons-material/Settings";
import SkipNextIcon from "@mui/icons-material/SkipNext";
import SkipPreviousIcon from "@mui/icons-material/SkipPrevious";
import SyncIcon from "@mui/icons-material/Sync";
import VerifiedUser from "@mui/icons-material/VerifiedUser";
import WarningIcon from "@mui/icons-material/Warning";
import * as React from "react";
import styled from "styled-components";
import images from "../lib/images";
import { colors, fontSizes, spacing } from "../typedefs/styled";
import Flex from "./Flex";
import ApplicationsIcon from "./NavIcons/ApplicationsIcon";
import ClusterDiscoveryIcon from "./NavIcons/ClusterDiscoveryIcon";
import ClustersIcon from "./NavIcons/ClustersIcon";
import DeliveryIcon from "./NavIcons/DeliveryIcon";
import DocsIcon from "./NavIcons/DocsIcon";
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

import ResumeIcon from "./Sync/ResumeIcon";
import Text from "./Text";

export enum IconType {
  Account,
  AddIcon,
  ApplicationsIcon,
  ArrowDownwardRoundedIcon,
  ArrowDropDownIcon,
  ArrowUpwardIcon,
  ArrowUpwardRoundedIcon,
  CallMade,
  CallReceived,
  CheckCircleIcon,
  CheckMark,
  ClearIcon,
  ClusterDiscoveryIcon,
  ClustersIcon,
  DeleteIcon,
  DeliveryIcon,
  DocsIcon,
  EditIcon,
  ErrorIcon,
  ExploreIcon,
  ExternalTab,
  FailedIcon,
  FileCopyIcon,
  FilterIcon,
  FindInPage,
  FluxIcon,
  FluxIconHover,
  GitOpsRunIcon,
  GitOpsSetsIcon,
  HourglassFullIcon,
  ImageAutomationIcon,
  InfoIcon,
  KeyboardArrowDownIcon,
  KeyboardArrowRightIcon,
  LogoutIcon,
  NavigateBeforeIcon,
  NavigateNextIcon,
  NotificationsIcon,
  PauseIcon,
  PendingActionIcon,
  PipelinesIcon,
  PlayIcon,
  PoliciesIcon,
  Policy,
  PolicyConfigsIcon,
  ReconcileIcon,
  Remove,
  RemoveCircleIcon,
  ResumeIcon,
  SaveAltIcon,
  SearchIcon,
  SecretsIcon,
  SettingsIcon,
  SkipNextIcon,
  SkipPreviousIcon,
  SourcesIcon,
  SuccessIcon,
  SuspendedIcon,
  SyncIcon,
  TemplatesIcon,
  TerraformIcon,
  VerifiedUser,
  WarningIcon,
  WorkspacesIcon,
}

type Props = {
  className?: string;
  type: IconType;
  color?: keyof typeof colors;
  text?: string;
  size: keyof typeof spacing;
  fontSize?: keyof typeof fontSizes;
};

function getIcon(i: IconType) {
  switch (i) {
    case IconType.Account:
      return PersonIcon;

    case IconType.AddIcon:
      return AddIcon;

    case IconType.ApplicationsIcon:
      return ApplicationsIcon;

    case IconType.ArrowDownwardRoundedIcon:
      return ArrowDownwardRoundedIcon;

    case IconType.ArrowDropDownIcon:
      return ArrowDropDownIcon;

    case IconType.ArrowUpwardIcon:
      return ArrowUpwardIcon;

    case IconType.ArrowUpwardRoundedIcon:
      return ArrowUpwardRoundedIcon;

    case IconType.CallMade:
      return CallMade;

    case IconType.CallReceived:
      return CallReceived;

    case IconType.CheckCircleIcon:
      return CheckCircleIcon;

    case IconType.CheckMark:
      return CheckCircleIcon;

    case IconType.ClearIcon:
      return ClearIcon;

    case IconType.ClusterDiscoveryIcon:
      return ClusterDiscoveryIcon;

    case IconType.ClustersIcon:
      return ClustersIcon;

    case IconType.DeleteIcon:
      return DeleteIcon;

    case IconType.DeliveryIcon:
      return DeliveryIcon;

    case IconType.DocsIcon:
      return DocsIcon;

    case IconType.EditIcon:
      return EditIcon;

    case IconType.ErrorIcon:
      return ErrorIcon;

    case IconType.ExploreIcon:
      return ExploreIcon;

    case IconType.ExternalTab:
      return LaunchIcon;

    case IconType.FailedIcon:
      return ErrorIcon;

    case IconType.FileCopyIcon:
      return FileCopyIcon;

    case IconType.FilterIcon:
      return FilterIcon;

    case IconType.FindInPage:
      return FindInPage;

    case IconType.FluxIcon:
      return FluxIcon;

    case IconType.GitOpsRunIcon:
      return GitOpsRunIcon;

    case IconType.GitOpsSetsIcon:
      return GitOpsSetsIcon;

    case IconType.HourglassFullIcon:
      return HourglassFullIcon;

    case IconType.ImageAutomationIcon:
      return ImageAutomationIcon;

    case IconType.InfoIcon:
      return InfoIcon;

    case IconType.KeyboardArrowDownIcon:
      return KeyboardArrowDownIcon;

    case IconType.KeyboardArrowRightIcon:
      return KeyboardArrowRightIcon;

    case IconType.LogoutIcon:
      return LogoutIcon;

    case IconType.NavigateBeforeIcon:
      return NavigateBeforeIcon;

    case IconType.NavigateNextIcon:
      return NavigateNextIcon;

    case IconType.NotificationsIcon:
      return NotificationsIcon;

    case IconType.PauseIcon:
      return PauseIcon;

    case IconType.PendingActionIcon:
      return () => <img src={images.pendingAction} />;

    case IconType.PipelinesIcon:
      return PipelinesIcon;

    case IconType.PlayIcon:
      return PlayIcon;

    case IconType.PoliciesIcon:
      return () => <PoliciesIcon filled={false} />;

    case IconType.Policy:
      return () => <PoliciesIcon filled />;

    case IconType.PolicyConfigsIcon:
      return PolicyConfigsIcon;

    case IconType.ReconcileIcon:
      return () => <img src={images.reconcileSrc} />;

    case IconType.Remove:
      return Remove;

    case IconType.RemoveCircleIcon:
      return RemoveCircleIcon;

    case IconType.ResumeIcon:
      return ResumeIcon;

    case IconType.SaveAltIcon:
      return SaveAltIcon;

    case IconType.SearchIcon:
      return SearchIcon;

    case IconType.SecretsIcon:
      return SecretsIcon;

    case IconType.SettingsIcon:
      return SettingsIcon;

    case IconType.SkipNextIcon:
      return SkipNextIcon;

    case IconType.SkipPreviousIcon:
      return SkipPreviousIcon;

    case IconType.SourcesIcon:
      return SourcesIcon;

    case IconType.SuccessIcon:
      return () => <img src={images.successSrc} />;

    case IconType.SuspendedIcon:
      return () => <img src={images.suspendedSrc} />;

    case IconType.SyncIcon:
      return SyncIcon;

    case IconType.TemplatesIcon:
      return TemplatesIcon;

    case IconType.TerraformIcon:
      return TerraformIcon;

    case IconType.VerifiedUser:
      return VerifiedUser;

    case IconType.WarningIcon:
      return WarningIcon;

    case IconType.WorkspacesIcon:
      return WorkspacesIcon;

    default:
      break;
  }
}

function Icon({ className, type, text, color, fontSize }: Props) {
  return (
    <Flex align className={className}>
      {React.createElement(getIcon(type) || "span")}
      {text && (
        <Text color={color} bold size={fontSize}>
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
    font-size: ${(props) => props.theme.fontSizes[props.fontSize as any]};
  }

  img {
    width: ${(props) => props.theme.spacing[props.size as any]};
  }
`;
