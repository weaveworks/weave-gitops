import * as React from "react";
import styled from "styled-components";
import Flex from "./Flex";
import Icon, { IconType } from "./Icon";
import Spacer from "./Spacer";
import Text from "./Text";

type StatusProps = {
  ok: boolean;
  msg: string;
  error: boolean;
};

function PageStatus({ ok, msg, error }: StatusProps) {
  return (
    <div className={`page-status${error ? " error-border" : ""}`}>
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
export default styled(PageStatus)`
  .page-status {
    max-width: 45%;
    padding: ${(props) => props.theme.spacing.small};
    color: ${(props) => props.theme.colors.neutral30};
    &.error-border {
      border: 1px solid ${(props) => props.theme.colors.neutral20};
      border-radius: 10px;
    }
  }
`;
