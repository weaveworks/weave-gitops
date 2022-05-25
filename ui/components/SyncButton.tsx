import * as React from "react";
import styled from "styled-components";
import Button, { IconButton } from "./Button";
import Flex from "./Flex";
import Icon, { IconType } from "./Icon";

type Props = {
  className?: string;
  loading?: boolean;
  onClick: (opts: { withSource: boolean }) => void;
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
  overflow: hidden;
  background: white;
  height: ${(props) => (props.open ? "40px" : "0px")};
  transition: height 0.25s ease-in;
`;

function SyncButton({ className, loading, onClick }: Props) {
  const [open, setOpen] = React.useState(false);
  return (
    <Flex column start className={className}>
      <Flex style={{ position: "relative" }}>
        <Button
          loading={loading}
          variant="outlined"
          onClick={() => onClick({ withSource: true })}
        >
          Sync
        </Button>
        <ArrowDropDown variant="outlined" onClick={() => setOpen(!open)}>
          <Icon type={IconType.ArrowDropDownIcon} size="base" />
        </ArrowDropDown>
      </Flex>
      <DropDown open={open}>
        <Button
          variant="outlined"
          color="primary"
          onClick={() => onClick({ withSource: false })}
        >
          Sync Without Source
        </Button>
      </DropDown>
    </Flex>
  );
}

export default styled(SyncButton).attrs({ className: SyncButton.name })``;
