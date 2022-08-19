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

function CheckboxActions({ className, checked }: Props) {
  //TODO add multi sync

  const makeSuspendReqs = () => {
    return checked.reduce((array, item) => {
      array.push({
        kind: item.kind,
        name: item.name,
        namespace: item.namespace,
        clusterName: item.clusterName,
      });
      return array;
    }, []);
  };

  const suspend = (suspend) =>
    useToggleSuspend(
      {
        objects: makeSuspendReqs(),
        suspend: suspend,
      },
      checked[0].kind
    );
  return (
    <Flex start align className={className}>
      <Button onClick={() => suspend(false).mutateAsync()}>
        <Icon type={IconType.PlayIcon} size="medium" />
      </Button>
      <Button onClick={() => suspend(true).mutateAsync()}>
        <Icon type={IconType.PauseIcon} size="medium" />
      </Button>
    </Flex>
  );
}

export default styled(CheckboxActions).attrs({
  className: CheckboxActions.name,
})``;
