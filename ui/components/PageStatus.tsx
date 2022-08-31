import * as React from "react";
import styled from "styled-components";
import { Condition } from "../lib/api/core/types.pb";
import { colors } from "../typedefs/styled.d";
import Flex from "./Flex";
import Icon, { IconType } from "./Icon";
import { computeMessage, computeReady, ReadyType } from "./KubeStatusIndicator";
import Spacer from "./Spacer";
import Text from "./Text";

type StatusProps = {
  conditions: Condition[];
  suspended: boolean;
  className?: string;
};

function PageStatus({ conditions, suspended, className }: StatusProps) {
  const msg = suspended ? "Suspended" : computeMessage(conditions);

  let iconType: IconType;
  let iconColor: keyof typeof colors;
  if (suspended) {
    iconType = IconType.SuspendedIcon;
    iconColor = "suspended";
  } else {
    const ok = computeReady(conditions);
    switch (ok) {
      case ReadyType.Reconciling:
        iconType = IconType.ReconcileIcon;
        iconColor = "primary";
        break;
      case ReadyType.Ready:
        iconType = IconType.CheckCircleIcon;
        iconColor = "success";
        break;
      case ReadyType.NotReady:
        iconType = IconType.FailedIcon;
        iconColor = "alert";
        break;
    }
  }

  return (
    <Flex align className={className}>
      <Icon type={iconType} color={iconColor} size="medium" />
      <Spacer padding="xs" />
      <Text color="neutral30">{msg}</Text>
    </Flex>
  );
}
export default styled(PageStatus).attrs({ className: PageStatus.name })`
  color: ${(props) => props.theme.colors.neutral30};
`;
