import * as React from "react";
import styled from "styled-components";
import { Condition } from "../lib/api/core/types.pb";
import Flex from "./Flex";
import Icon, { IconType } from "./Icon";
import { computeMessage, computeReady } from "./KubeStatusIndicator";
import Spacer from "./Spacer";
import Text from "./Text";

type StatusProps = {
  conditions: Condition[];
  error: boolean;
  className?: string;
};

function PageStatus({ conditions, error, className }: StatusProps) {
  const ok = computeReady(conditions);
  const msg = computeMessage(conditions);
  return (
    <div className={`${className}${!ok || error ? " error-border" : ""}`}>
      <Flex align>
        <Icon
          type={ok ? IconType.CheckCircleIcon : IconType.FailedIcon}
          color={ok ? "success" : "alert"}
          size="medium"
        />
        <Spacer padding="xs" />
        <Text color="neutral30">{msg}</Text>
      </Flex>
    </div>
  );
}
export default styled(PageStatus).attrs({ className: PageStatus.name })`
  max-width: 45%;
  margin-right: ${(props) => props.theme.spacing.medium};
  padding: ${(props) => props.theme.spacing.small};
  color: ${(props) => props.theme.colors.neutral30};
  &.error-border {
    border: 1px solid ${(props) => props.theme.colors.neutral20};
    border-radius: 10px;
  }
`;
