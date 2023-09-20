import { Tooltip } from "@material-ui/core";
import React from "react";
import { useGetObject } from "../hooks/objects";
import { Condition, Kind, ObjectRef } from "../lib/api/core/types.pb";
import {
  Automation,
  GitRepository,
  OCIRepository,
  Source,
} from "../lib/objects";
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

const checkVerifiable = (sourceRef: ObjectRef): boolean => {
  //guard against an undefined or non-verifiable obj (as of right now anything that's not a git or oci repo)
  const { name, namespace, kind } = sourceRef;
  const undefinedRef = !name || !namespace || !kind;
  return (
    (kind === Kind.GitRepository || kind === Kind.OCIRepository) &&
    !undefinedRef
  );
};

const VerifiedStatus = ({
  source,
}: {
  source?: VerifiableSource;
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

export const SourceIsVerifiedStatus: React.FC<{
  automation?: Automation;
  sourceRef?: ObjectRef;
  source?: Source;
}> = ({ automation, sourceRef, source }): JSX.Element => {
  const isVerifiable = source
    ? checkVerifiable({
        name: source.name,
        namespace: source.namespace,
        clusterName: source.clusterName,
        kind: source?.type,
      })
    : checkVerifiable(sourceRef || {});
  if (!isVerifiable) return <Flex>-</Flex>;

  if (source) return <VerifiedStatus source={source as VerifiableSource} />;
  const {
    name = "",
    namespace = automation.namespace,
    kind = "",
    clusterName = automation.clusterName,
  } = sourceRef || {};
  const { data: verifiable } = useGetObject<VerifiableSource>(
    name,
    namespace,
    Kind[kind],
    clusterName
  );

  return <VerifiedStatus source={verifiable} />;
};
