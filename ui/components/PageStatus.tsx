import * as React from "react";
import styled from "styled-components";
import { Condition } from "../lib/api/core/types.pb";
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
  const ok = suspended ? false : computeReady(conditions);
  const msg = suspended ? "Suspended" : computeMessage(conditions);
  return (
    <Flex align className={className}>
      <Icon
        type={
          suspended
            ? IconType.SuspendedIcon
            : ok
            ? ok === ReadyType.Reconciling
              ? IconType.ReconcileIcon
              : IconType.CheckCircleIcon
            : IconType.FailedIcon
        }
        color={ok ? "success" : "alert"}
        size="medium"
      />
      <Spacer padding="xs" />
      <Text color="neutral30">{msg}</Text>
    </Flex>
  );
}
export default styled(PageStatus).attrs({ className: PageStatus.name })`
  color: ${(props) => props.theme.colors.neutral30};
`;
