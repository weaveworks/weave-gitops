import React from "react";
import styled from "styled-components";
import { FluxObject } from "../lib/objects";
import Flex from "./Flex";
import { HealthStatusType } from "./HealthCheckStatusIndicator";
import Icon, { IconType } from "./Icon";
import Text from "./Text";

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
    <Flex wide align gap="16">
      {health.unhealthy > 0 ? (
        <>
          <Icon
            type={IconType.FailedIcon}
            color="alertOriginal"
            size="medium"
          />
          <Text color="neutral30">{`${health.unhealthy} workload(s) are failing health checks`}</Text>
        </>
      ) : (
        <>
          <Icon
            type={IconType.CheckCircleIcon}
            color="successOriginal"
            size="medium"
          />
          <Text color="neutral30">All workloads are passing health checks</Text>
        </>
      )}
    </Flex>
  );
};

export default styled(HealthCheckAgg).attrs({})``;
