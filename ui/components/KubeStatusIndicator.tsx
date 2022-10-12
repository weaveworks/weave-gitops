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
  short?: boolean;
  suspended?: boolean;
};
export enum ReadyType {
  Ready = "Ready",
  NotReady = "Not Ready",
  Reconciling = "Reconciling",
}

export function computeReady(conditions: Condition[]): ReadyType {
  if (!conditions?.length) return undefined;
  const readyCondition =
    _.find(conditions, (c) => c.type === "Ready") ||
    _.find(conditions, (c) => c.type === "Available");
  if (readyCondition) {
    if (readyCondition.status === "True") return ReadyType.Ready;
    if (
      readyCondition.status === "Unknown" &&
      readyCondition.reason === "Progressing"
    )
      return ReadyType.Reconciling;
    return ReadyType.NotReady;
  }

  if (_.find(conditions, (c) => c.status === "False"))
    return ReadyType.NotReady;
  return ReadyType.Ready;
}

export function computeMessage(conditions: Condition[]) {
  if (!conditions?.length) return undefined;
  const readyCondition =
    _.find(conditions, (c) => c.type === "Ready") ||
    _.find(conditions, (c) => c.type === "Available");
  if (readyCondition) return readyCondition.message;

  const falseCondition = _.find(conditions, (c) => c.status === "False");
  if (falseCondition) return falseCondition.message;
  return conditions[0].message;
}

function KubeStatusIndicator({
  className,
  conditions,
  short,
  suspended,
}: Props) {
  let readyText;
  let icon;
  let iconColor;
  if (suspended) {
    readyText = "Suspended";
    icon = IconType.SuspendedIcon;
  } else {
    const ready = computeReady(conditions);
    if (ready === ReadyType.Reconciling) {
      readyText = ReadyType.Reconciling;
      icon = IconType.ReconcileIcon;
      iconColor = "primary";
    } else if (ready === ReadyType.Ready) {
      readyText = ReadyType.Ready;
      icon = IconType.CheckCircleIcon;
      iconColor = "success";
    } else {
      readyText = ReadyType.NotReady;
      icon = IconType.FailedIcon;
      iconColor = "alert";
    }
  }

  let text = computeMessage(conditions);
  if (short || suspended) text = readyText;

  return (
    <Flex start className={className} align>
      <Icon size="base" type={icon} color={iconColor} text={text} />
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
