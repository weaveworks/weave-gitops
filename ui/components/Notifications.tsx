import { Box, Collapse } from "@material-ui/core";
import Alert from "@material-ui/lab/Alert";
import React, { FC, useContext } from "react";
import styled from "styled-components";
import { NotificationData, Notification } from "../contexts/NotificationsContext";
import Icon, { IconType } from "./Icon";
import Text from "./Text";

const BoxWrapper = styled(Box)<{ severity: string }>`
  div[class*="MuiAlert-root"] {
    width: auto;
    margin-bottom: ${(props) => props.theme.spacing.base};
    border-radius: ${(props) => props.theme.spacing.xs};
  }
  div[class*="MuiAlert-action"] {
    display: inline;
    color: ${(props) => {
      if (props.severity === "error") return props.theme.colors.alertLight;
      else if (props.severity === "warning")
        return props.theme.colors.feedbackLight;
      else if (props.severity === "success")
        return props.theme.colors.successLight;
      else return "transparent";
    }};
    svg {
      fill: ${(props) => {
        if (props.severity === "error") return props.theme.colors.alertMedium;
        else if (props.severity === "warning")
          return props.theme.colors.feedbackMedium;
        else if (props.severity === "success")
          return props.theme.colors.successMedium;
        else return "transparent";
      }};
    }
  }
  div[class*="MuiAlert-icon"] {
    svg[class*="MuiSvgIcon-root"] {
      display: none;
    }
  }
  div[class*="MuiAlert-message"] {
    display: flex;
    justify-content: center;
    align-items: center;
    svg {
      margin-right: ${(props) => props.theme.spacing.xs};
    }
  }
  div[class*="MuiAlert-standardError"] {
    background-color: ${(props) => props.theme.colors.alertLight};
  }
  div[class*="MuiAlert-standardSuccess"] {
    background-color: ${(props) => props.theme.colors.successLight};
  }
  div[class*="MuiAlert-standardWarning"] {
    background-color: ${(props) => props.theme.colors.alertLight};
  }
`;

const Notifications: FC<{ notifications: NotificationData[] }> = ({
  notifications,
}) => {
  const { setNotifications } = useContext(Notification);

  const handleDelete = (n: NotificationData) =>
    setNotifications(
      notifications.filter(
        (notif) =>
          (n.message.text !== notif.message.text ||
            n.message.component !== notif.message.component) &&
          n.severity !== notif.severity
      )
    );

  const getIcon = (severity?: string) => {
    switch (severity) {
      case "error":
        return <Icon type={IconType.ErrorIcon} size="medium" />;
      case "success":
        return <Icon type={IconType.SuccessIcon} size="medium" />;
      case "warning":
        return <Icon type={IconType.WarningIcon} size="medium" />;
      default:
        return;
    }
  };

  const notificationAlert = (n: NotificationData, index: number) => {
    return (
      <BoxWrapper key={index} severity={n?.severity || ""}>
        <Collapse in={true}>
          <Alert severity={n?.severity} onClose={() => handleDelete(n)}>
            {getIcon(n?.severity)}
            <Text color="black">{n?.message.text}</Text> {n?.message.component}
          </Alert>
        </Collapse>
      </BoxWrapper>
    );
  };

  return (
    <Box style={{ width: "100%" }}>
      {notifications.map((n, index) => {
        return (
          (n?.message.text || n?.message.component) &&
          notificationAlert(n, index)
        );
      })}
    </Box>
  );
};

export default Notifications;
