import AddIcon from "@material-ui/icons/Add";
import ArrowDownwardRoundedIcon from "@material-ui/icons/ArrowDownwardRounded";
import ArrowDropDownIcon from "@material-ui/icons/ArrowDropDown";
import ArrowUpwardIcon from "@material-ui/icons/ArrowUpward";
import ArrowUpwardRoundedIcon from "@material-ui/icons/ArrowUpwardRounded";
import CallMade from "@material-ui/icons/CallMade";
import CallReceived from "@material-ui/icons/CallReceived";
import CheckCircleIcon from "@material-ui/icons/CheckCircle";
import ClearIcon from "@material-ui/icons/Clear";
import DeleteIcon from "@material-ui/icons/Delete";
import EditIcon from "@material-ui/icons/Edit";
import ErrorIcon from "@material-ui/icons/Error";
import LogoutIcon from "@material-ui/icons/ExitToApp";
import FileCopyIcon from "@material-ui/icons/FileCopyOutlined";
import FilterIcon from "@material-ui/icons/FilterList";
import FindInPage from "@material-ui/icons/FindInPage";
import HourglassFullIcon from "@material-ui/icons/HourglassFull";
import InfoIcon from "@material-ui/icons/Info";
import KeyboardArrowDownIcon from "@material-ui/icons/KeyboardArrowDown";
import KeyboardArrowRightIcon from "@material-ui/icons/KeyboardArrowRight";
import LaunchIcon from "@material-ui/icons/Launch";
import NavigateBeforeIcon from "@material-ui/icons/NavigateBefore";
import NavigateNextIcon from "@material-ui/icons/NavigateNext";
import PauseIcon from "@material-ui/icons/Pause";
import PersonIcon from "@material-ui/icons/Person";
import PlayIcon from "@material-ui/icons/PlayArrow";
import Policy from "@material-ui/icons/Policy";
import Remove from "@material-ui/icons/Remove";
import RemoveCircleIcon from "@material-ui/icons/RemoveCircle";
import SaveAltIcon from "@material-ui/icons/SaveAlt";
import SearchIcon from "@material-ui/icons/Search";
import SettingsIcon from "@material-ui/icons/Settings";
import SkipNextIcon from "@material-ui/icons/SkipNext";
import SkipPreviousIcon from "@material-ui/icons/SkipPrevious";
import SyncIcon from "@material-ui/icons/Sync";
import VerifiedUser from "@material-ui/icons/VerifiedUser";
import WarningIcon from "@material-ui/icons/Warning";
import * as React from "react";
import styled from "styled-components";
import images from "../lib/images";
// eslint-disable-next-line
import { colors, fontSizes, spacing } from "../typedefs/styled";
import Flex from "./Flex";
import ApplicationsIcon from "./NavIcons/ApplicationsIcon";
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
import ClusaterDiscoveryIcon from "./NavIcons/ClusterDiscoveryIcon";
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
  ClusterDiscoveryIcon,
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
      return PoliciesIcon;

    case IconType.Policy:
      return Policy;

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

    case IconType.ClusterDiscoveryIcon:
      return ClusaterDiscoveryIcon;

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
