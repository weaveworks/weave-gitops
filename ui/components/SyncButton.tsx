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

const ArrowDropDown = styled(IconButton)`
  &.MuiButton-outlined {
    border-color: ${(props) => props.theme.colors.neutral20};
    border-left: 0px;
  }
  &.MuiButton-root {
    border-radius: 0;
    min-width: 0;
    height: initial;
    padding: 7px 0px;
  }
  &.MuiButton-text {
    padding: 0;
  }
`;

const DropDown = styled(Flex)`
  position: absolute;
  overflow: hidden;
  background: white;
  height: ${(props) => (props.open ? "100%" : "0px")};
  transition: height 0.2s ease-in;
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
        <Button
          disabled={disabled}
          loading={loading}
          variant="outlined"
          onClick={() => onClick({ withSource: true })}
          style={{ marginRight: 0 }}
        >
          Sync
        </Button>
        {arrowDropDown}
      </Flex>
      <DropDown open={open}>
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
