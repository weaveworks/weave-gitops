import * as React from "react";
import styled from "styled-components";
import { colors } from "../typedefs/styled.d";
import Flex from "./Flex";
import Icon, { IconType } from "./Icon";
import Text from "./Text";

export enum HealthStatusType {
  Progressing = "Progressing",
  Healthy = "Healthy",
  Unhealthy = "Unhealthy",
  Unknown = "Unknown",
}

type IndicatorInfo = {
  icon: IconType;
  color: keyof typeof colors;
  type: HealthStatusType;
};

export const getIndicatorInfo = (
  healthType: HealthStatusType,
): IndicatorInfo => {
  switch (healthType) {
    case HealthStatusType.Unhealthy:
      return {
        icon: IconType.FailedIcon,
        color: "alertOriginal",
        type: HealthStatusType.Unhealthy,
      };
    case HealthStatusType.Healthy:
      return {
        icon: IconType.CheckCircleIcon,
        color: "successOriginal",
        type: HealthStatusType.Healthy,
      };
    case HealthStatusType.Progressing:
      return {
        type: HealthStatusType.Progressing,
        icon: IconType.ReconcileIcon,
        color: "primary",
      };
    default:
      return {
        icon: IconType.SuspendedIcon,
        color: "feedbackOriginal",
        type: HealthStatusType.Unknown,
      };
  }
};

function HealthCheckStatusIndicator({
  className,
  health,
}: {
  className?: string;
  health: { message?: string; status?: string };
}) {
  const { type, color, icon } = getIndicatorInfo(
    HealthStatusType[health.status],
  );

  const text = health.message ? health.message : type;

  return (
    <Flex start className={className} align>
      <Icon size="base" type={icon} color={color} text={text} />
    </Flex>
  );
}

export default styled(HealthCheckStatusIndicator).attrs({
  className: HealthCheckStatusIndicator.name,
})`
  ${Icon} ${Text} {
    color: ${(props) => props.theme.colors.neutral40};
    font-weight: 400;
  }
`;
