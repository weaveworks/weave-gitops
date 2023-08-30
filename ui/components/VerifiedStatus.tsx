import { Tooltip } from "@material-ui/core";
import React from "react";
import { useGetObject } from "../hooks/objects";
import { Condition, Kind, ObjectRef } from "../lib/api/core/types.pb";
import { GitRepository, OCIRepository } from "../lib/objects";
import Flex from "./Flex";
import Icon, { IconType } from "./Icon";

type VerifiableSource = GitRepository | OCIRepository;

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
  a: VerifiableSource | undefined
): Condition | undefined =>
  a?.conditions?.find((condition) => condition.type === "SourceVerified");

export const VerifiedStatus = ({
  source,
}: {
  source: VerifiableSource;
}): JSX.Element => {
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
  const { name, namespace, kind, clusterName } = sourceRef;

  const objKind = kind && Kind[kind];
  //can sourceRef actually return undefined stuff?! Typescript says it can.
  const undefinedRef = !name || !namespace || !kind;
  if (
    (kind !== Kind.GitRepository && kind !== Kind.OCIRepository) ||
    undefinedRef
  )
    return <Flex>-</Flex>;

  const { data: source } = useGetObject<VerifiableSource>(
    name,
    namespace,
    objKind,
    clusterName || ""
  );

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
