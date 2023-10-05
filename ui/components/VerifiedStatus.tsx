import React from "react";
import { Tooltip } from "@material-ui/core";
import { Condition, ObjectRef } from "../lib/api/core/types.pb";
import { GitRepository, OCIRepository } from "../lib/objects";
import { useListSources } from "../hooks/sources";
import Icon, { IconType } from "./Icon";

export interface VerifiableSource {
  isVerifiable: boolean;
  conditions: Condition[];
}

const getVerifiedStatusColor = (status?: string) => {
  let color;
  if (status === "True") {
    color = "successOriginal";
  } else if (status === "False") {
    color = "alertOriginal";
  } else if (!status) {
    color = "feedbackOriginal";
  }
  return color;
};

export const findVerificationCondition = (
  a: VerifiableSource
): Condition | undefined =>
  a?.conditions?.find((condition) => condition.type === "SourceVerified");

export const VerifiedStatus = ({
  source,
}: {
  source: VerifiableSource;
}): JSX.Element | null => {
  if (!source.isVerifiable) return null;

  const condition = findVerificationCondition(source);
  const color = getVerifiedStatusColor(condition?.status);

  return (
    <Tooltip title={condition?.message || "pending verification"}>
      <div>
        <Icon type={IconType.VerifiedUser} color={color} size="base" />
      </div>
    </Tooltip>
  );
};

export const SourceIsVerifiedStatus: React.FC<{ sourceRef: ObjectRef }> = ({
  sourceRef,
}): JSX.Element | null => {
  const { data: sources } = useListSources();
  const currentSource = sources?.result.find(
    (source) => sourceRef?.name === source.name
  ) as GitRepository | OCIRepository | undefined;

  if (!currentSource?.isVerifiable) return null;

  const condition = findVerificationCondition(currentSource);
  const color = getVerifiedStatusColor(condition?.status);

  return (
    <Tooltip title={condition?.message || "pending verification"}>
      <div>
        <Icon type={IconType.VerifiedUser} color={color} size="base" />
      </div>
    </Tooltip>
  );
};
