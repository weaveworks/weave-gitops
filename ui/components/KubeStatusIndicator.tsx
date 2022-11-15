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

export enum ReadyStatusValue {
  True = "True",
  False = "False",
  Unknown = "Unknown",
}

export function computeReady(conditions: Condition[]): ReadyType {
  if (!conditions?.length) return undefined;
  const readyCondition =
    _.find(conditions, (c) => c.type === "Ready") ||
    _.find(conditions, (c) => c.type === "Available");
  if (readyCondition) {
    if (readyCondition.status === ReadyStatusValue.True) {
      return ReadyType.Ready;
    }

    if (
      readyCondition.status === ReadyStatusValue.Unknown &&
      readyCondition.reason === "Progressing"
    ) {
      return ReadyType.Reconciling;
    }

    return ReadyType.NotReady;
  }

  if (_.find(conditions, (c) => c.status === ReadyStatusValue.False)) {
    return ReadyType.NotReady;
  }

  return ReadyType.Ready;
}

export function computeMessage(conditions: Condition[]) {
  if (!conditions?.length) {
    return undefined;
  }

  const readyCondition =
    _.find(conditions, (c) => c.type === "Ready") ||
    _.find(conditions, (c) => c.type === "Available");

  if (readyCondition) {
    return readyCondition.message;
  }

  const falseCondition = _.find(
    conditions,
    (c) => c.status === ReadyStatusValue.False
  );

  if (falseCondition) {
    return falseCondition.message;
  }

  return conditions[0].message;
}

type SpecialObject = "Daemonset";

interface DaemonSetStatus {
  currentNumberScheduled: number;
  desiredNumberScheduled: number;
  numberMisscheduled: number;
  numberReady: number;
  numberUnavailable: number;
  observedGeneration: number;
  updatedNumberScheduled: number;
}

const NotReady: Condition[] = [
  {
    type: ReadyType.Ready,
    status: ReadyStatusValue.False,
    message: "Not Ready",
  },
];
const Ready: Condition[] = [
  { type: ReadyType.Ready, status: ReadyStatusValue.True, message: "Ready" },
];
const Unknown: Condition[] = [
  { type: ReadyType.Ready, status: ReadyStatusValue.Unknown },
];

// Certain objects to not have a status.conditions key, so we generate those conditions
// and feed it into the `KubeStatusIndicator` to keep the public API consistent.
export function createSyntheticConditions(
  kind: SpecialObject,
  // This will eventually be a union type when we add another special object.
  // Example: DaemonSetStatus | CoolObjectStatus | ...
  status: DaemonSetStatus
): Condition[] {
  switch (kind) {
    case "Daemonset":
      if (status.numberReady === status.desiredNumberScheduled) {
        return Ready;
      }

      return NotReady;

    default:
      return Unknown;
  }
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
