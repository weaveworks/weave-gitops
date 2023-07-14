import { Tooltip } from "@material-ui/core";
import _ from "lodash";
import * as React from "react";
import { useLocation } from "react-router-dom";
import styled from "styled-components";
import { useSyncFluxObject } from "../hooks/automations";
import { useToggleSuspend } from "../hooks/flux";
import { ObjectRef } from "../lib/api/core/types.pb";
import { V2Routes } from "../lib/types";
import Button from "./Button";
import Flex from "./Flex";
import Icon, { IconType } from "./Icon";
import SyncButton from "./SyncButton";

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

const DefaultSync: React.FC<{ reqObjects: ObjectRef[] }> = ({ reqObjects }) => {
  const defaultSync = useSyncFluxObject(reqObjects);
  const location = useLocation();
  const noSource = {
    [V2Routes.Sources]: true,
  };
  return (
    <SyncButton
      disabled={reqObjects[0] ? false : true}
      loading={defaultSync.isLoading}
      onClick={(opts) => defaultSync.mutateAsync(opts)}
      hideDropdown={noSource[location.pathname]}
    />
  );
};

const DefaultSuspend: React.FC<{
  reqObjects: ObjectRef[];
  suspend: boolean;
}> = ({ reqObjects, suspend }) => {
  function createDefaultSuspendHandler(
    reqObjects: ObjectRef[],
    suspend: boolean
  ) {
    const result = useToggleSuspend(
      {
        objects: reqObjects,
        suspend: suspend,
      },
      reqObjects[0]?.kind === "HelmRelease" ||
        reqObjects[0]?.kind === "Kustomization"
        ? "automations"
        : "sources"
    );
    return () => result.mutateAsync();
  }

  return (
    <Tooltip
      title={suspend ? "Suspend Selected" : "Resume Selected"}
      placement="top"
    >
      <div>
        <Button
          disabled={reqObjects[0] ? false : true}
          onClick={createDefaultSuspendHandler(reqObjects, suspend)}
        >
          <Icon
            type={suspend ? IconType.PauseIcon : IconType.PlayIcon}
            size="medium"
          />
        </Button>
      </div>
    </Tooltip>
  );
};

type Props = {
  className?: string;
  checked?: string[];
  rows?: any[];
};

function CheckboxActions({ className, checked = [], rows = [] }: Props) {
  const [reqObjects, setReqObjects] = React.useState([]);
  const hasChecked = checked.length > 0;

  React.useEffect(() => {
    if (hasChecked && rows.length) setReqObjects(makeObjects(checked, rows));
    else setReqObjects([]);
  }, [checked, rows]);

  return (
    <Flex start align className={className} gap="8">
      <DefaultSync reqObjects={reqObjects} />
      <DefaultSuspend reqObjects={reqObjects} suspend={true} />
      <DefaultSuspend reqObjects={reqObjects} suspend={false} />
    </Flex>
  );
}

export default styled(CheckboxActions).attrs({
  className: CheckboxActions.name,
})`
  margin-right: 8px;
`;
