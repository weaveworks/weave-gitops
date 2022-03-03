import _ from "lodash";
import * as React from "react";
import styled from "styled-components";
import { Condition } from "../lib/api/core/types.pb";
import Flex from "./Flex";
import Icon, { IconType } from "./Icon";
import Text from "./Text";

type Props = {
  className?: string;
  conditions: Condition[];
};

export function computeReady(conditions: Condition[]): boolean {
  const ready =
    _.find(conditions, { type: "Ready" }) ||
    // Deployment conditions work slightly differently;
    // they show "Available" instead of 'Ready'
    _.find(conditions, { type: "Available" });
  return ready?.status == "True";
}

export function computeMessage(conditions: Condition[]) {
  const readyCondition = _.find(conditions, (c) => c.type === "Ready");

  return readyCondition.message;
}

function KubeStatusIndicator({ className, conditions }: Props) {
  const ready = computeReady(conditions);
  const readyText = ready ? "Ready" : computeMessage(conditions);
  const color = ready ? "success" : "alert";

  return (
    <Flex start className={className} align>
      <Icon color={color} size="base" type={IconType.Circle} text={readyText} />
    </Flex>
  );
}

export default styled(KubeStatusIndicator).attrs({
  className: KubeStatusIndicator.name,
})`
  ${Icon} ${Text} {
    color: ${(props) => props.theme.colors.black};
    font-weight: 400;
  }
`;
