import {
  IconButton,
  ListItemIcon,
  Menu,
  MenuItem,
  Tooltip,
} from "@material-ui/core";
import * as React from "react";
import { useNavigate } from "react-router-dom-v5-compat";
import styled from "styled-components";
import { Auth } from "../contexts/AuthContext";
import { V2Routes } from "../lib/types";
import Icon, { IconType } from "./Icon";

const SettingsMenu = styled(Menu)`
  .MuiList-root {
    padding: ${(props) => props.theme.spacing.small};
  }
  .logout {
    justify-content: flex-end;
    .MuiListItemIcon-root {
      min-width: 0;
      color: ${(props) => props.theme.colors.black};
    }
    .MuiSvgIcon-root {
      padding-right: ${(props) => props.theme.spacing.xs};
    }
  }
`;

const PersonButton = styled(IconButton)<{ open: boolean }>`
  height: 40px;
  width: 40px;
  &.MuiIconButton-root {
    background-color: ${(props) => props.theme.colors.neutralGray};
    color: ${(props) => props.theme.colors.neutral30};
    :hover {
      background-color: ${(props) => props.theme.colors.white};
      color: ${(props) => props.theme.colors.primary10};
    }
  }
`;

function UserSettings({ className }: { className?: string }) {
  const navigate = useNavigate();
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
    <div className={className}>
      <Tooltip title="Account settings" enterDelay={500} enterNextDelay={500}>
        <PersonButton
          onClick={handleClick}
          aria-controls={open ? "account-menu" : undefined}
          aria-haspopup="true"
          aria-expanded={open ? "true" : undefined}
          open={open}
        >
          <Icon size="medium" type={IconType.Account} />
        </PersonButton>
      </Tooltip>
      <SettingsMenu
        anchorEl={anchorEl}
        id="account-menu"
        open={open}
        onClose={handleClose}
        onClick={handleClose}
        transformOrigin={{ horizontal: 150, vertical: -80 }}
      >
        <MenuItem disabled>Hello, {userInfo?.email}</MenuItem>
        <MenuItem onClick={() => navigate(V2Routes.Notifications)}>
          Notifications
        </MenuItem>
        <MenuItem className="logout" onClick={() => logOut()}>
          <ListItemIcon>
            <Icon type={IconType.LogoutIcon} size="base" />
          </ListItemIcon>
          Logout
        </MenuItem>
      </SettingsMenu>
    </div>
  );
}

export default styled(UserSettings)`
  padding-right: ${(props) => props.theme.spacing.small};
`;
