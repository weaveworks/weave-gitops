import * as React from "react";
import styled from "styled-components";
import { Condition } from "../lib/api/app/source.pb";
import { computeMessage, computeReady } from "../lib/utils";
import Flex from "./Flex";
import Icon, { IconType } from "./Icon";

type Props = {
  className?: string;
  conditions: Condition[];
};

function KubeStatusIndicator({ className, conditions }: Props) {
  const ready = computeReady(conditions) === "True";
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
})``;
