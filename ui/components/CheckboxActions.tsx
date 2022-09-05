import { Tooltip } from "@material-ui/core";
import _ from "lodash";
import * as React from "react";
import styled from "styled-components";
import { useSyncFluxObject } from "../hooks/automations";
import { useToggleSuspend } from "../hooks/flux";
import Button from "./Button";
import Flex from "./Flex";
import Icon, { IconType } from "./Icon";
import Spacer from "./Spacer";
import SyncButton from "./SyncButton";

export const makeObjects = (checked: string[], rows: any[]) => {
  const objects = [];
  checked.forEach((uid) => {
    const row = _.find(rows, (row) => {
      return uid === row.uid;
    });
    if (row)
      return objects.push({
        kind: row.kind,
        name: row.name,
        namespace: row.namespace,
        clusterName: row.clusterName,
      });
  });
  return objects;
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
  }, [checked, rows]);

  function createSuspendHandler(suspend: boolean) {
    const result = useToggleSuspend(
      {
        objects: reqObjects,
        suspend: suspend,
      },
      reqObjects[0] ? reqObjects[0].kind : ""
    );

    return () => result.mutateAsync();
  }

  const sync = useSyncFluxObject(reqObjects);

  return (
    <Flex start align className={className}>
      <SyncButton
        onClick={(opts) => sync.mutateAsync(opts)}
        loading={sync.isLoading}
        disabled={!hasChecked}
      />
      <Spacer padding="xxs" />
      <Tooltip title="Suspend Selected" placement="top">
        <div>
          <Button disabled={!hasChecked} onClick={createSuspendHandler(true)}>
            <Icon type={IconType.PauseIcon} size="medium" />
          </Button>
        </div>
      </Tooltip>
      <Spacer padding="xxs" />
      <Tooltip title="Resume Selected" placement="top">
        <div>
          <Button disabled={!hasChecked} onClick={createSuspendHandler(false)}>
            <Icon type={IconType.PlayIcon} size="medium" />
          </Button>
        </div>
      </Tooltip>
      <Spacer padding="xxs" />
    </Flex>
  );
}

export default styled(CheckboxActions).attrs({
  className: CheckboxActions.name,
})``;
