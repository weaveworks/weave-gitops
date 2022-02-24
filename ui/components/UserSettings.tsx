import {
  IconButton,
  ListItemIcon,
  Menu,
  MenuItem,
  Tooltip,
} from "@material-ui/core";
import * as React from "react";
import styled from "styled-components";
import { Auth } from "../contexts/AuthContext";
import Icon, { IconType } from "./Icon";

const UserAvatar = styled(Icon)`
  padding-right: ${(props) => props.theme.spacing.medium};
`;

const SettingsMenu = styled(Menu)`
  .MuiListItemIcon-root {
    min-width: 25px;
    color: ${(props) => props.theme.colors.black};
  }
`;

function UserSettings() {
  const [anchorEl, setAnchorEl] = React.useState(null);
  const { userInfo, logOut } = React.useContext(Auth);

  const open = Boolean(anchorEl);
  const handleClick = (event) => {
    setAnchorEl(event.currentTarget);
  };
  const handleClose = () => {
    setAnchorEl(null);
  };

  return (
    <>
      <Tooltip title="Account settings">
        <IconButton
          onClick={handleClick}
          aria-controls={open ? "account-menu" : undefined}
          aria-haspopup="true"
          aria-expanded={open ? "true" : undefined}
        >
          <UserAvatar size="xl" type={IconType.Account} color="white" />
        </IconButton>
      </Tooltip>
      <SettingsMenu
        anchorEl={anchorEl}
        id="account-menu"
        open={open}
        onClose={handleClose}
        onClick={handleClose}
        transformOrigin={{ horizontal: "right", vertical: "top" }}
      >
        <MenuItem>Hello, {userInfo?.email}</MenuItem>
        <MenuItem onClick={() => logOut()}>
          <ListItemIcon>
            <Icon type={IconType.LogoutIcon} size="base" />
          </ListItemIcon>
          Logout
        </MenuItem>
      </SettingsMenu>
    </>
  );
}

export default styled(UserSettings)``;
