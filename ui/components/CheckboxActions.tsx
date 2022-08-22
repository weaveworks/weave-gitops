import * as React from "react";
import styled from "styled-components";
import { useToggleSuspend } from "../hooks/flux";
import Button from "./Button";
import Flex from "./Flex";
import Icon, { IconType } from "./Icon";

type Props = {
  className?: string;
  checked?: any[];
};

function CheckboxActions({ className, checked = [] }: Props) {
  //TODO add multi sync

  const makeSuspendReqs = (arr: any[]) => {
    return arr.reduce((array, item) => {
      array.push({
        kind: item.kind,
        name: item.name,
        namespace: item.namespace,
        clusterName: item.clusterName,
      });
      return array;
    }, []);
  };

  const objects = makeSuspendReqs(checked);
  const resume = useToggleSuspend(
    {
      objects: objects,
      suspend: false,
    },
    objects[0] ? objects[0].kind : ""
  );
  const suspend = useToggleSuspend(
    {
      objects: objects,
      suspend: true,
    },
    objects[0] ? objects[0].kind : ""
  );

  return (
    <Flex start align className={className}>
      <Button
        loading={resume.isLoading}
        disabled={!checked[0]}
        onClick={() => resume.mutateAsync()}
      >
        <Icon type={IconType.PlayIcon} size="medium" />
      </Button>
      <Button
        loading={suspend.isLoading}
        disabled={!checked[0]}
        onClick={() => suspend.mutateAsync()}
      >
        <Icon type={IconType.PauseIcon} size="medium" />
      </Button>
    </Flex>
  );
}

export default styled(CheckboxActions).attrs({
  className: CheckboxActions.name,
})``;
