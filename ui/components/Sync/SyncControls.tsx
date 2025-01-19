import {
  FormControl,
  FormControlLabel,
  Radio,
  RadioGroup,
  Tooltip,
} from "@mui/material";
import { alpha } from "@mui/material/styles";
import React, { type JSX } from "react";
import styled, { keyframes } from "styled-components";
import { ThemeTypes } from "../../contexts/AppContext";
import Button from "../Button";
import CustomActions from "../CustomActions";
import Flex from "../Flex";
import Icon, { IconType } from "../Icon";
import Spacer from "../Spacer";

interface Props {
  className?: string;
  syncLoading?: boolean;
  syncDisabled?: boolean;
  suspendDisabled?: boolean;
  resumeDisabled?: boolean;
  hideSyncOptions?: boolean;
  hideSuspend?: boolean;
  tooltipSuffix?: string;
  customActions?: JSX.Element[];
  onSyncClick: (syncType: SyncType) => void;
  onSuspendClick?: () => void;
  onResumeClick?: () => void;
}

export enum SyncType {
  WithSource = "WithSource",
  WithoutSource = "WithoutSource",
}

const rotateAnimation = keyframes`
 0% { transform: rotate(0deg); }
 100% { transform: rotate(-360deg); }
`;

const SourceLabel = styled(FormControlLabel)`
  &.MuiFormControlLabel-root {
    margin-right: 0;
    margin-left: 0;
  }

  .MuiFormControlLabel-label {
    color: ${(props) =>
      props.theme.mode === ThemeTypes.Dark
        ? props.theme.colors.neutral30
        : props.theme.colors.neutral40};

    &.Mui-disabled {
      color: ${(props) =>
        props.theme.mode === ThemeTypes.Dark
          ? props.theme.colors.primary30
          : props.theme.colors.neutral20};
    }
  }

  .MuiTypography-root {
    margin-left: ${(props) => props.theme.spacing.xs};
  }
`;

export const IconButton = styled(Button)`
  &.MuiButton-root {
    border-radius: 50%;
    min-width: 32px;
    height: 32px;
    padding: 0;

    &.Mui-disabled {
      svg {
        fill: ${(props) =>
          props.theme.mode === ThemeTypes.Dark
            ? props.theme.colors.primary30
            : props.theme.colors.neutral20};
      }
    }

    &:hover {
      background-color: ${(props) =>
        props.theme.mode === ThemeTypes.Dark
          ? alpha(props.theme.colors.primary10, 0.2)
          : alpha(props.theme.colors.primary, 0.1)};
    }
  }
  &.MuiButton-text {
    padding: 0;

    &.Mui-disabled {
      color: ${(props) =>
        props.theme.mode === ThemeTypes.Dark
          ? props.theme.colors.primary30
          : props.theme.colors.neutral20};
      svg {
        fill: ${(props) =>
          props.theme.mode === ThemeTypes.Dark
            ? props.theme.colors.primary30
            : props.theme.colors.neutral20};
      }
    }
  }
`;

const SyncControls = ({
  className,
  syncLoading,
  syncDisabled,
  suspendDisabled,
  resumeDisabled,
  hideSyncOptions,
  hideSuspend,
  tooltipSuffix = "",
  customActions,
  onSyncClick,
  onSuspendClick,
  onResumeClick,
}: Props) => {
  const [syncType, setSyncType] = React.useState(
    hideSyncOptions ? SyncType.WithoutSource : SyncType.WithSource,
  );

  const handleSyncTypeChange = (value: SyncType) => {
    setSyncType(value);
  };

  const disableSyncButtons = syncDisabled || syncLoading;

  return (
    <Flex wide start align className={className}>
      {customActions && (
        <>
          <CustomActions actions={customActions} />
          <Spacer padding="xxs" />
        </>
      )}
      <Button
        className="sync-icon-button"
        variant="text"
        loading={false}
        disabled={disableSyncButtons}
        startIcon={
          <Icon
            type={IconType.SyncIcon}
            size="medium"
            className={syncLoading ? "rotate-icon" : ""}
          />
        }
      >
        Sync
      </Button>
      <Spacer padding="xxs" />
      {!hideSyncOptions && (
        <>
          <FormControl>
            <RadioGroup
              data-testid="sync-options"
              row
              aria-labelledby="source-radio-buttons-group-label"
              name="source-radio-buttons-group"
              value={syncType}
              onChange={(event) => {
                handleSyncTypeChange(event.target.value as SyncType);
              }}
            >
              <SourceLabel
                value={SyncType.WithSource}
                control={<Radio />}
                label="with Source"
                disabled={disableSyncButtons}
              />
              <Spacer padding="xxs" />
              <SourceLabel
                value={SyncType.WithoutSource}
                control={<Radio />}
                label="without Source"
                disabled={disableSyncButtons}
              />
            </RadioGroup>
          </FormControl>
          <Spacer padding="xxs" />
        </>
      )}
      <Tooltip title={`Sync${tooltipSuffix}`} placement="top">
        <div>
          <IconButton
            data-testid="sync-button"
            disabled={disableSyncButtons}
            onClick={() => onSyncClick(syncType)}
            size="large"
          >
            <Icon type={IconType.PlayIcon} size="medium" />
          </IconButton>
        </div>
      </Tooltip>
      {!hideSuspend && (
        <>
          <Spacer padding="xxs" />
          <Tooltip title={`Suspend${tooltipSuffix}`} placement="top">
            <div>
              <IconButton
                data-testid="suspend-button"
                disabled={suspendDisabled}
                onClick={onSuspendClick}
                size="large"
              >
                <Icon type={IconType.PauseIcon} size="medium" />
              </IconButton>
            </div>
          </Tooltip>
          <Spacer padding="xxs" />
          <Tooltip title={`Resume${tooltipSuffix}`} placement="top">
            <div>
              <IconButton
                data-testid="resume-button"
                disabled={resumeDisabled}
                onClick={onResumeClick}
                size="large"
              >
                <Icon
                  type={IconType.ResumeIcon}
                  size="medium"
                  color="primary10"
                />
              </IconButton>
            </div>
          </Tooltip>
        </>
      )}
    </Flex>
  );
};

export default styled(SyncControls)`
  .sync-icon-button {
    text-transform: uppercase;
    pointer-events: none;
  }

  .rotate-icon {
    color: ${(props) => props.theme.colors.primary10};
    animation: 1s linear infinite ${rotateAnimation};
  }
`;
