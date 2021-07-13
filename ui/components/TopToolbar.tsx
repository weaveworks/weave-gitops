import { IconButton, Menu, MenuItem } from "@material-ui/core";
import * as React from "react";
import styled from "styled-components";
import useAuth from "../hooks/auth";
import Flex from "./Flex";
import Icon, { IconType } from "./Icon";
import Logo from "./Logo";

type Props = {
  className?: string;
};

const UserAvatar = styled(Icon)`
  padding-right: ${(props) => props.theme.spacing.medium};
`;

function TopToolbar({ className }: Props) {
  const { user, logout } = useAuth();
  const [anchor, setAnchor] = React.useState(null);

  const handleMenuOpen = (ev: React.MouseEvent<HTMLElement>) => {
    setAnchor(ev.currentTarget);
  };

  const handleClose = () => {
    setAnchor(null);
  };

  return (
    <Flex align between className={className}>
      <Logo />
      <div>
        <IconButton onClick={handleMenuOpen}>
          {user && (
            <UserAvatar size="xl" type={IconType.Account} color="white" />
          )}
        </IconButton>
        <Menu
          id="user-menu"
          anchorEl={anchor}
          open={Boolean(anchor)}
          onClose={handleClose}
        >
          <MenuItem>{(user || {}).email}</MenuItem>
          <MenuItem
            onClick={() => {
              handleClose();
              logout();
            }}
          >
            Logout
          </MenuItem>
        </Menu>
      </div>
    </Flex>
  );
}

export default styled(TopToolbar)`
  display: flex;
  padding: 8px 0;
  background-color: ${(props) => props.theme.colors.primary};
  width: 100%;
  height: 112px;
  flex: 0 1 auto;

  ${Logo} img {
    width: 136px;
    height: 42px;
  }
`;
