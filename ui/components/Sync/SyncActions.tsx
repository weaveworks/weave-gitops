import React from "react";
import styled from "styled-components";
import { useSyncFluxObject } from "../../hooks/automations";
import { useToggleSuspend } from "../../hooks/flux";
import { Kind } from "../../lib/api/core/types.pb";
import SyncControls, { SyncType } from "./SyncControls";

interface Props {
  name?: string;
  namespace?: string;
  clusterName?: string;
  kind?: Kind;
  suspended?: boolean;
  hideSyncOptions?: boolean;
  customActions?: JSX.Element[];
  className?: string;
}

const SyncActions = ({
  name,
  namespace,
  clusterName,
  kind,
  suspended,
  hideSyncOptions,
  customActions,
  className,
}: Props) => {
  const sync = useSyncFluxObject([
    {
      name,
      namespace,
      clusterName,
      kind: kind,
    },
  ]);

  const syncHandler = (syncType: SyncType) => {
    sync.mutateAsync({ withSource: syncType === SyncType.WithSource });
  };

  const objects = [
    {
      name,
      namespace,
      clusterName,
      kind: kind,
    },
  ];

  const suspend = useToggleSuspend(
    {
      objects: objects,
      suspend: true,
    },
    "object"
  );

  const resume = useToggleSuspend(
    {
      objects: objects,
      suspend: false,
    },
    "object"
  );

  return (
    <SyncControls
      className={className}
      hideSyncOptions={hideSyncOptions}
      syncLoading={sync.isLoading}
      syncDisabled={suspended}
      suspendDisabled={suspend.isLoading || suspended}
      resumeDisabled={resume.isLoading || !suspended}
      customActions={customActions}
      onSyncClick={syncHandler}
      onSuspendClick={() => suspend.mutateAsync()}
      onResumeClick={() => resume.mutateAsync()}
    />
  );
};

export default styled(SyncActions)`
  width: 50%;
  min-width: fit-content;
`;
