import _ from "lodash";
import * as React from "react";
import styled from "styled-components";
import { useToggleSuspend } from "../hooks/flux";
import { FluxObjectKind } from "../lib/api/core/types.pb";
import Button from "./Button";
import Flex from "./Flex";
import Icon, { IconType } from "./Icon";
import Spacer from "./Spacer";

export type Unique = {
  kind: FluxObjectKind;
  name: string;
  namespace: string;
  clusterName: string;
};

export const makeUniques = (arr: any[]) => {
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

export const compareUniques = (uniques: Unique[], check: Unique) => {
  let equal: boolean = false;
  uniques.map((unique, index) => {
    //_.isEqual deeply compares stuff instead of using stinky memory addresses
    if (_.isEqual(unique, check)) {
      return (equal = true);
    }
  });
  return equal;
};

export const removeUnique = (uniques: Unique[], item: Unique) => {
  _.remove(uniques, (unique) => _.isEqual(unique, item));
  return uniques;
};

type Props = {
  className?: string;
  checked?: Unique[];
};

function CheckboxActions({ className, checked = [] }: Props) {
  //TODO add multi sync

  const objects = checked;
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
        loading={suspend.isLoading}
        disabled={!checked[0]}
        onClick={() => suspend.mutateAsync()}
      >
        <Icon type={IconType.PauseIcon} size="medium" />
      </Button>
      <Spacer padding="xxs" />
      <Button
        loading={resume.isLoading}
        disabled={!checked[0]}
        onClick={() => resume.mutateAsync()}
      >
        <Icon type={IconType.PlayIcon} size="medium" />
      </Button>
    </Flex>
  );
}

export default styled(CheckboxActions).attrs({
  className: CheckboxActions.name,
})``;
