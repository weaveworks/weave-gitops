import React, { type JSX } from "react";
import styled from "styled-components";
import { useSyncFluxObject } from "../../hooks/automations";
import { useToggleSuspend } from "../../hooks/flux";
import { Kind } from "../../lib/api/core/types.pb";
import SuspendMessageModal from "./SuspendMessageModal";
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
  const [suspendMessageModalOpen, setSuspendMessageModalOpen] =
    React.useState(false);
  const [suspendMessage, setSuspendMessage] = React.useState("");
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
      comment: suspendMessage,
    },
    "object",
  );

  const resume = useToggleSuspend(
    {
      objects: objects,
      suspend: false,
      comment: "",
    },
    "object",
  );

  return (
    <>
      <SyncControls
        className={className}
        hideSyncOptions={hideSyncOptions}
        syncLoading={sync.isPending}
        syncDisabled={suspended}
        suspendDisabled={suspend.isPending || suspended}
        resumeDisabled={resume.isPending || !suspended}
        customActions={customActions}
        onSyncClick={syncHandler}
        onSuspendClick={() =>
          setSuspendMessageModalOpen(!suspendMessageModalOpen)
        }
        onResumeClick={() => resume.mutateAsync()}
      />
      <SuspendMessageModal
        open={suspendMessageModalOpen}
        onCloseModal={setSuspendMessageModalOpen}
        suspend={suspend}
        setSuspendMessage={setSuspendMessage}
        suspendMessage={suspendMessage}
      />
    </>
  );
};

export default styled(SyncActions)`
  width: 50%;
  min-width: fit-content;
`;
