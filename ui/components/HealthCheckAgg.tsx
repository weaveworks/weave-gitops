import React from "react";
import { FluxObject } from "../lib/objects";
import Flex from "./Flex";
import { HealthStatusType } from "./HealthCheckStatusIndicator";
import Icon, { IconType } from "./Icon";

export interface AggHealth {
  healthy: number;
  unhealthy: number;
  progressing: number;
  NA: number;
}

export function computeAggHealthCheck(objects: FluxObject[]): AggHealth {
  const healthAgg: AggHealth = {
    healthy: 0,
    unhealthy: 0,
    NA: 0,
    progressing: 0,
  };
  objects.forEach(({ health }) => {
    switch (health.status) {
      case HealthStatusType.Healthy:
        healthAgg.healthy += 1;
        break;
      case HealthStatusType.Unhealthy:
        healthAgg.unhealthy += 1;
        break;
      case HealthStatusType.Progressing:
        healthAgg.progressing += 1;
        break;
      default:
        healthAgg.NA += 1;
        break;
    }
  });
  return healthAgg;
}
interface Prop {
  health: AggHealth;
}
const HealthCheckAgg = ({ health }: Prop) => {
  return (
    <Flex wide gap="14">
      {health.progressing > 0 && (
        <Icon
          size="base"
          type={IconType.ReconcileIcon}
          color="primary"
          text={`Progressing: ${health.progressing}`}
        />
      )}
      <Icon
        size="base"
        type={IconType.FailedIcon}
        color="alertOriginal"
        text={`Unhealthy: ${health.unhealthy}`}
      />
      <Icon
        size="base"
        type={IconType.CheckCircleIcon}
        color="successOriginal"
        text={`Healthy: ${health.healthy}`}
      />
      <Icon
        size="base"
        type={IconType.SuspendedIcon}
        color="feedbackOriginal"
        text={`NA: ${health.NA}`}
      />
    </Flex>
  );
};

export default HealthCheckAgg;
