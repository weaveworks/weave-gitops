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
  NotReady = "NotReady",
  Reconciling = "Reconciling",
}

export function computeReady(conditions: Condition[]): string {
  if (
    _.find(conditions, (c) => c.type === "Ready") ||
    _.find(conditions, (c) => c.type === "Available")
  ) {
    return _.find(conditions, (c) => c.status === "Unknown") &&
      _.find(conditions, (c) => c.reason === "Progressing")
      ? ReadyType.Reconciling
      : ReadyType.Ready;
  }
  return undefined;
}

export function computeMessage(conditions: Condition[]) {
  const readyCondition =
    _.find(conditions, (c) => c.type === "Ready") ||
    _.find(conditions, (c) => c.type === "Available");

  return readyCondition ? readyCondition.message : "unknown error";
}

function KubeStatusIndicator({
  className,
  conditions,
  short,
  suspended,
}: Props) {
  let readyText;
  let icon;
  if (suspended) {
    readyText = "Suspended";
    icon = IconType.SuspendedIcon;
  } else {
    const ready = computeReady(conditions);
    if (ready) {
      if (ready === ReadyType.Reconciling) {
        readyText = ReadyType.Reconciling;
        icon = IconType.ReconcileIcon;
      } else {
        readyText = ReadyType.Ready;
        icon = IconType.SuccessIcon;
      }
    } else {
      readyText = ReadyType.NotReady;
      icon = IconType.FailedIcon;
    }
  }

  let text = computeMessage(conditions);
  if (short || suspended) text = readyText;

  return (
    <Flex start className={className} align>
      <Icon size="base" type={icon} text={text} />
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
