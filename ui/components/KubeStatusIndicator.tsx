import _ from "lodash";
import * as React from "react";
import styled from "styled-components";
import { Condition } from "../lib/api/core/types.pb";
import { colors } from "../typedefs/styled.d";
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
  PendingAction = "PendingAction",
  Suspended = "Suspended",
  None = "None",
}

export enum ReadyStatusValue {
  True = "True",
  False = "False",
  Unknown = "Unknown",
  None = "None",
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

    if (readyCondition.status === ReadyStatusValue.Unknown) {
      if (readyCondition.reason === "Progressing") return ReadyType.Reconciling;
      if (readyCondition.reason === "TerraformPlannedWithChanges")
        return ReadyType.PendingAction;
    }

    if (readyCondition.status === ReadyStatusValue.None) return ReadyType.None;

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

type IndicatorInfo = {
  icon: IconType;
  color: keyof typeof colors;
  type: ReadyType;
};

export const getIndicatorInfo = (
  suspended: boolean,
  conditions: Condition[]
): IndicatorInfo => {
  if (suspended)
    return {
      icon: IconType.SuspendedIcon,
      color: "feedbackOriginal",
      type: ReadyType.Suspended,
    };
  const ready = computeReady(conditions);
  if (ready === ReadyType.Reconciling)
    return {
      type: ReadyType.Reconciling,
      icon: IconType.ReconcileIcon,
      color: "primary",
    };
  if (ready === ReadyType.PendingAction)
    return {
      type: ReadyType.PendingAction,
      icon: IconType.PendingActionIcon,
      color: "feedbackOriginal",
    };
  if (ready === ReadyType.Ready)
    return {
      type: ReadyType.Ready,
      icon: IconType.CheckCircleIcon,
      color: "successOriginal",
    };
  if (ready === ReadyType.None)
    return {
      type: ReadyType.None,
      icon: IconType.RemoveCircleIcon,
      color: "neutral20",
    };
  return {
    type: ReadyType.NotReady,
    icon: IconType.FailedIcon,
    color: "alertOriginal",
  };
};

export type SpecialObject = "DaemonSet";

interface DaemonSetStatus {
  currentNumberScheduled: number;
  desiredNumberScheduled: number;
  numberMisscheduled: number;
  numberReady: number;
  numberUnavailable: number;
  observedGeneration: number;
  updatedNumberScheduled: number;
}

const NotReady: Condition = {
  type: ReadyType.Ready,
  status: ReadyStatusValue.False,
  message: "Not Ready",
};

const Ready: Condition = {
  type: ReadyType.Ready,
  status: ReadyStatusValue.True,
  message: "Ready",
};

const Unknown: Condition = {
  type: ReadyType.Ready,
  status: ReadyStatusValue.Unknown,
  message: "Unknown",
};

// Certain objects to not have a status.conditions key, so we generate those conditions
// and feed it into the `KubeStatusIndicator` to keep the public API consistent.
export function createSyntheticCondition(
  kind: SpecialObject,
  // This will eventually be a union type when we add another special object.
  // Example: DaemonSetStatus | CoolObjectStatus | ...
  status: DaemonSetStatus
): Condition {
  switch (kind) {
    case "DaemonSet":
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
  const { type, color, icon } = getIndicatorInfo(suspended, conditions);

  let text = computeMessage(conditions);
  if (short || suspended) text = type;

  return (
    <Flex start className={className} align>
      <Icon size="base" type={icon} color={color} text={text} />
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
