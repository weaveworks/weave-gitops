import {
  IconButton,
  ListItemIcon,
  Menu,
  MenuItem,
  Tooltip,
} from "@mui/material";
import * as React from "react";
import { useNavigate } from "react-router";
import styled from "styled-components";
import { Auth } from "../contexts/AuthContext";
import { V2Routes } from "../lib/types";
import DarkModeSwitch from "./DarkModeSwitch";
import Flex from "./Flex";
import Icon, { IconType } from "./Icon";
import Link from "./Link";

const SettingsMenu = styled(Menu)`
  .MuiPaper-root {
    background: ${(props) => props.theme.colors.whiteToPrimary};
  }
  .MuiList-root {
    padding: ${(props) => props.theme.spacing.small};
  }
  .logout {
    justify-content: flex-end;
    .MuiListItemIcon-root {
      min-width: 0;
      color: ${(props) => props.theme.colors.neutral40};
    }
    .MuiSvgIcon-root {
      padding-right: ${(props) => props.theme.spacing.xs};
    }
  }
`;

const PersonButton = styled(IconButton)<{ open?: boolean }>`
  &.MuiIconButton-root {
    padding: ${(props) => props.theme.spacing.xs};
    background-color: ${(props) => props.theme.colors.neutralGray};
    color: ${(props) => props.theme.colors.neutral30};
    &:hover {
      background-color: ${(props) => props.theme.colors.blueWithOpacity};
    }
  }
`;

type Props = {
  className?: string;
  darkModeEnabled?: boolean;
};

function UserSettings({ className, darkModeEnabled = true }: Props) {
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
    <Flex className={className} gap="8" align>
      <DarkModeSwitch darkModeEnabled={darkModeEnabled} />
      <Tooltip title="Docs" enterDelay={500} enterNextDelay={500}>
        <Link
          as={PersonButton}
          href="https://docs.gitops.weaveworks.org/"
          target="_blank"
          rel="noreferrer"
        >
          <Icon size="medium" type={IconType.FindInPage} />
        </Link>
      </Tooltip>
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
        transformOrigin={{ horizontal: 90, vertical: -10 }}
      >
        <MenuItem onClick={() => navigate(V2Routes.UserInfo)}>
          Hello, {userInfo?.id}
        </MenuItem>
        <MenuItem className="logout" onClick={() => logOut()}>
          <ListItemIcon>
            <Icon type={IconType.LogoutIcon} size="base" />
          </ListItemIcon>
          Logout
        </MenuItem>
      </SettingsMenu>
    </Flex>
  );
}

export default styled(UserSettings)``;
