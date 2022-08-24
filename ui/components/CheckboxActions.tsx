import _ from "lodash";
import * as React from "react";
import styled from "styled-components";
import { useToggleSuspend } from "../hooks/flux";
import Button from "./Button";
import Flex from "./Flex";
import Icon, { IconType } from "./Icon";
import Spacer from "./Spacer";

const makeObjects = (checked: string[], rows: any[]) => {
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
  //TODO add multi sync
  const [suspendReqs, setSuspendReqs] = React.useState([]);

  React.useEffect(() => {
    if (checked.length && rows.length)
      setSuspendReqs(makeObjects(checked, rows));
  }, [checked, rows]);

  const resume = useToggleSuspend(
    {
      objects: suspendReqs,
      suspend: false,
    },
    suspendReqs[0] ? suspendReqs[0].kind : ""
  );
  const suspend = useToggleSuspend(
    {
      objects: suspendReqs,
      suspend: true,
    },
    suspendReqs[0] ? suspendReqs[0].kind : ""
  );

  return (
    <Flex start align className={className}>
      <Button disabled={!checked[0]} onClick={() => suspend.mutateAsync()}>
        <Icon type={IconType.PauseIcon} size="medium" />
      </Button>
      <Spacer padding="xxs" />
      <Button disabled={!checked[0]} onClick={() => resume.mutateAsync()}>
        <Icon type={IconType.PlayIcon} size="medium" />
      </Button>
    </Flex>
  );
}

export default styled(CheckboxActions).attrs({
  className: CheckboxActions.name,
})``;
