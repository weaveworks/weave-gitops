import _ from "lodash";
import * as React from "react";
import { useLocation } from "react-router";
import styled from "styled-components";
import { useSyncFluxObject } from "../../hooks/automations";
import { useToggleSuspend } from "../../hooks/flux";
import { ObjectRef } from "../../lib/api/core/types.pb";
import { V2Routes } from "../../lib/types";
import SuspendMessageModal from "./SuspendMessageModal";
import SyncControls, { SyncType } from "./SyncControls";

export const makeObjects = (checked: string[], rows: any[]): ObjectRef[] => {
  const objects = [];
  checked.forEach((uid) => {
    const row = _.find(rows, (row) => {
      return uid === row.uid;
    });
    if (row)
      return objects.push({
        kind: row.type,
        name: row.name,
        namespace: row.namespace,
        clusterName: row.clusterName,
      });
  });
  return objects;
};

const noSource = {
  [V2Routes.Sources]: true,
  [V2Routes.ImageRepositories]: true,
  [V2Routes.ImageUpdates]: true,
};

function createSuspendHandler(
  reqObjects: ObjectRef[],
  suspend: boolean,
  suspendMessage: string,
) {
  const result = useToggleSuspend(
    {
      objects: reqObjects,
      suspend: suspend,
      comment: suspendMessage,
    },
    reqObjects[0]?.kind === "HelmRelease" ||
      reqObjects[0]?.kind === "Kustomization"
      ? "automations"
      : "sources",
  );
  return result;
}

type Props = {
  className?: string;
  checked?: string[];
  rows?: any[];
};

function CheckboxActions({ className, checked = [], rows = [] }: Props) {
  const [reqObjects, setReqObjects] = React.useState([]);
  const [suspendMessageModalOpen, setSuspendMessageModalOpen] =
    React.useState(false);
  const [suspendMessage, setSuspendMessage] = React.useState("");
  const location = useLocation();

  React.useEffect(() => {
    if (checked.length > 0 && rows.length)
      setReqObjects(makeObjects(checked, rows));
    else setReqObjects([]);
  }, [checked, rows]);

  const sync = useSyncFluxObject(reqObjects);

  const syncHandler = (syncType: SyncType) => {
    sync.mutateAsync({ withSource: syncType === SyncType.WithSource });
  };

  const disableButtons = !reqObjects[0];

  return (
    <>
      <SyncControls
        className={className}
        hideSyncOptions={noSource[location.pathname]}
        syncLoading={sync.isPending}
        syncDisabled={disableButtons}
        suspendDisabled={disableButtons}
        resumeDisabled={disableButtons}
        tooltipSuffix=" selected"
        onSyncClick={syncHandler}
        onSuspendClick={() =>
          setSuspendMessageModalOpen(!suspendMessageModalOpen)
        }
        onResumeClick={
          createSuspendHandler(reqObjects, false, suspendMessage).mutateAsync
        }
      />
      <SuspendMessageModal
        open={suspendMessageModalOpen}
        onCloseModal={setSuspendMessageModalOpen}
        suspend={createSuspendHandler(reqObjects, true, suspendMessage)}
        setSuspendMessage={setSuspendMessage}
        suspendMessage={suspendMessage}
      />
    </>
  );
}

export default styled(CheckboxActions).attrs({
  className: CheckboxActions.name,
})`
  width: 50%;
  min-width: fit-content;
  margin-right: 8px;
`;
