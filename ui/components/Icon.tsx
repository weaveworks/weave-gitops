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
import Flex from "./Flex";
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
      return () => (
        <svg
          width="24"
          height="24"
          viewBox="0 0 24 24"
          fill="none"
          xmlns="http://www.w3.org/2000/svg"
          className="sources"
        >
          <path
            d="M10.7 14.3L8.4 12L10.7 9.7L10 9L7 12L10 15L10.7 14.3ZM13.3 14.3L15.6 12L13.3 9.7L14 9L17 12L14 15L13.3 14.3V14.3Z"
            fill="#1a1a1a"
          />
          <rect
            x="4.5"
            y="4.5"
            width="15"
            height="15"
            rx="7.5"
            stroke="#1A1A1A"
          />
        </svg>
      );

    case IconType.ImageAutomationIcon:
      return () => (
        <svg
          width="24"
          height="24"
          viewBox="0 0 24 24"
          fill="none"
          xmlns="http://www.w3.org/2000/svg"
        >
          <path
            d="M12 9L15 12H12.75V18.75H11.25V12H9L12 9ZM18.75 18.75H14.25V17.2575H18.75V6.735H5.25V17.2575H9.75V18.75H5.25C4.425 18.75 3.75 18.075 3.75 17.25V6.75C3.75 5.925 4.425 5.25 5.25 5.25H18.75C19.575 5.25 20.25 5.925 20.25 6.75V17.25C20.25 18.075 19.575 18.75 18.75 18.75ZM12 9L15 12H12.75V18.75H11.25V12H9L12 9ZM18.75 18.75H14.25V17.2575H18.75V6.735H5.25V17.2575H9.75V18.75H5.25C4.425 18.75 3.75 18.075 3.75 17.25V6.75C3.75 5.925 4.425 5.25 5.25 5.25H18.75C19.575 5.25 20.25 5.925 20.25 6.75V17.25C20.25 18.075 19.575 18.75 18.75 18.75Z"
            fill="#1A1A1A"
            className="image-automation"
          />
        </svg>
      );

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
    fill: ${(props) => props.theme.colors[props.color as any]} !important;
    height: ${(props) => props.theme.spacing[props.size as any]};
    width: ${(props) => props.theme.spacing[props.size as any]};
    path {
      &.image-automation {
        fill: ${(props) => props.theme.colors[props.color as any]} !important;
      }
    }
    &.sources {
      fill: none !important;
      path {
        fill: ${(props) => props.theme.colors[props.color as any]} !important;
      }
      rect {
        stroke: ${(props) => props.theme.colors[props.color as any]} !important;
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
