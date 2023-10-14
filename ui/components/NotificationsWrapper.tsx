import { Alert } from "@material-ui/lab";
import React, { FC, useContext, useEffect } from "react";
import styled from "styled-components";

import {
  NotificationData,
  Notification,
} from "../contexts/NotificationsContext";
import { ListError } from "../lib/api/core/core.pb";
import { AlertListErrors } from "./AlertListErrors";
import Flex from "./Flex";
import Icon, { IconType } from "./Icon";
import Notifications from "./Notifications";

const ENTITLEMENT_ERROR =
  "No entitlement was found for Weave GitOps Enterprise. Please contact sales@weave.works.";

export const Title = styled.h2`
  margin-top: 0px;
`;

interface Props {
  errors?: ListError[];
  warningMsg?: string;
  versionEntitlement?: string;
}

export const NotificationsWrapperOSS: FC<Props> = ({
  children,
  errors,
  warningMsg,
  versionEntitlement,
}) => {
  const { notifications, setNotifications } = useContext(Notification);

  const WarningWrapper = styled(Alert)`
    background: ${(props) => props.theme.colors.feedbackLight} !important;
    margin-bottom: ${(props) => props.theme.spacing.small};
    height: 50px;
    border-radius: ${(props) => props.theme.spacing.xs} !important;
    color: ${(props) => props.theme.colors.black} !important;
    display: flex !important;
    align-items: center;
    padding-right: 0 !important;
    padding-left: 0 !important;
    .MuiAlert-icon {
      margin-left: ${(props) => props.theme.spacing.base} !important;
    }
  `;
  useEffect(() => {
    if (versionEntitlement && versionEntitlement !== "") {
      setNotifications([
        {
          message: {
            text: versionEntitlement,
          },
          severity: "warning",
        } as NotificationData,
      ]);
    }
  }, [setNotifications]);

  const topNotifications = notifications?.filter(
    (n: any) => n.display !== "bottom" && n.message.text !== ENTITLEMENT_ERROR
  );
  const bottomNotifications = notifications?.filter(
    (n: any) => n.display === "bottom"
  );
  return (
    <div style={{ width: "100%" }}>
      ERoRRRRRRRRRRRRRRRRRRRRRRRR
      {errors && (
        <AlertListErrors
          errors={errors.filter((error) => error.message !== ENTITLEMENT_ERROR)}
        />
      )}
      {!!warningMsg && (
        <WarningWrapper
          severity="warning"
          iconMapping={{
            warning: <Icon type={IconType.WarningIcon} size="medium" />,
          }}
        >
          <span>{warningMsg}</span>
        </WarningWrapper>
      )}
      <Notifications notifications={topNotifications || []} />
      {children}
      {bottomNotifications && !!bottomNotifications.length && (
        <Flex wide style={{ paddingTop: "16px" }}>
          <Notifications notifications={bottomNotifications} />
        </Flex>
      )}
    </div>
  );
};
