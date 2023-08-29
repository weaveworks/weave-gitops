import * as React from "react";
import styled from "styled-components";
import Button, { IconButton } from "./Button";
import Flex from "./Flex";
import Icon, { IconType } from "./Icon";

type Props = {
  className?: string;
  loading?: boolean;
  disabled?: boolean;
  onClick: (opts: { withSource: boolean }) => void;
  hideDropdown?: boolean;
};

export const ArrowDropDown = styled(IconButton)`
  &.MuiButton-outlined {
    //2px = MUI radius
    border-radius: 0 2px 2px 0;
  }
  &.MuiButton-outlinedPrimary {
    border-color: ${(props) => props.theme.colors.neutral20};
  }
  &.MuiButton-root {
    min-width: 0;
    height: 32px;
    padding: 8px 4px;
  }
  &.MuiButton-text {
    padding: 0;
  }
`;

const Sync = styled(Button)<{ $hideDropdown: boolean }>`
  &.MuiButton-outlined {
    margin-right: 0;
    ${(props) =>
      !props.$hideDropdown && `border-radius: 2px 0 0 2px; border-right: none`}
  }
  &.MuiButton-outlined.Mui-disabled {
    border-right: none;
  }
`;

export const DropDown = styled(Flex)`
  position: absolute;
  overflow: hidden;
  background: ${(props) => props.theme.colors.white};
  height: ${(props) => (props.open ? "100%" : "0px")};
  transition-property: height;
  transition-duration: 0.2s;
  transition-timing-function: ease-in-out;
  z-index: 1;
`;

function SyncButton({
  className,
  loading,
  disabled,
  onClick,
  hideDropdown = false,
}: Props) {
  const [open, setOpen] = React.useState(false);
  let arrowDropDown;
  if (hideDropdown == false) {
    arrowDropDown = (
      <ArrowDropDown
        variant="outlined"
        color="primary"
        onClick={() => setOpen(!open)}
        disabled={disabled}
      >
        <Icon type={IconType.ArrowDropDownIcon} size="base" />
      </ArrowDropDown>
    );
  } else {
    arrowDropDown = <></>;
  }
  return (
    <div
      className={className}
      style={{ position: "relative", display: open ? "block" : "inline-block" }}
    >
      <Flex>
        <Sync
          disabled={disabled || loading}
          loading={loading}
          variant="outlined"
          onClick={() => onClick({ withSource: true })}
          //$ - transient prop that is not passed to DOM https://styled-components.com/docs/api#transient-props
          $hideDropdown={hideDropdown}
        >
          Sync
        </Sync>
        {arrowDropDown}
      </Flex>
      <DropDown open={open} absolute={true}>
        <Button
          variant="outlined"
          color="primary"
          onClick={() => onClick({ withSource: false })}
          style={{ whiteSpace: "nowrap" }}
        >
          Sync Without Source
        </Button>
      </DropDown>
    </div>
  );
}

export default styled(SyncButton).attrs({ className: SyncButton.name })``;
